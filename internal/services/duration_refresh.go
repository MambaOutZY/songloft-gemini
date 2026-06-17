package services

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"songloft/internal/database/sqlc"
	"songloft/internal/models"
)

type DurationRefreshStatus = string

const (
	DurationRefreshIdle       DurationRefreshStatus = "idle"
	DurationRefreshRunning    DurationRefreshStatus = "running"
	DurationRefreshCancelling DurationRefreshStatus = "cancelling"
	DurationRefreshDone       DurationRefreshStatus = "done"
	DurationRefreshCancelled  DurationRefreshStatus = "cancelled"
	DurationRefreshFailed     DurationRefreshStatus = "failed"
)

type DurationRefreshProgress struct {
	Status    string `json:"status"`
	Total     int    `json:"total"`
	Processed int    `json:"processed"`
	Failed    int    `json:"failed"`
}

type DurationRefresher struct {
	mu       sync.Mutex
	progress DurationRefreshProgress
	cancelFn context.CancelFunc

	listSongs  func(ctx context.Context) ([]sqlc.ListSongsNeedingDurationRow, error)
	updateDur  func(ctx context.Context, id int64, duration float64) error
	resolveURL func(ctx context.Context, song *models.Song) (string, error)
	extractor  *MetadataExtractor
}

func NewDurationRefresher(
	listSongs func(ctx context.Context) ([]sqlc.ListSongsNeedingDurationRow, error),
	updateDur func(ctx context.Context, id int64, duration float64) error,
	resolveURL func(ctx context.Context, song *models.Song) (string, error),
	extractor *MetadataExtractor,
) *DurationRefresher {
	return &DurationRefresher{
		progress:   DurationRefreshProgress{Status: DurationRefreshIdle},
		listSongs:  listSongs,
		updateDur:  updateDur,
		resolveURL: resolveURL,
		extractor:  extractor,
	}
}

func (d *DurationRefresher) Start() error {
	d.mu.Lock()
	if d.progress.Status == DurationRefreshRunning || d.progress.Status == DurationRefreshCancelling {
		d.mu.Unlock()
		return ErrDurationRefreshRunning
	}
	ctx, cancel := context.WithCancel(context.Background())
	d.cancelFn = cancel
	d.progress = DurationRefreshProgress{Status: DurationRefreshRunning}
	d.mu.Unlock()

	go d.run(ctx)
	return nil
}

func (d *DurationRefresher) GetProgress() DurationRefreshProgress {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.progress
}

func (d *DurationRefresher) Cancel() {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.cancelFn != nil && d.progress.Status == DurationRefreshRunning {
		d.cancelFn()
		d.progress.Status = DurationRefreshCancelling
	}
}

func (d *DurationRefresher) run(ctx context.Context) {
	defer func() {
		d.mu.Lock()
		switch d.progress.Status {
		case DurationRefreshRunning:
			d.progress.Status = DurationRefreshDone
		case DurationRefreshCancelling:
			d.progress.Status = DurationRefreshCancelled
		}
		d.cancelFn = nil
		d.mu.Unlock()
	}()

	songs, err := d.listSongs(ctx)
	if err != nil {
		slog.Warn("duration refresh: list songs failed", "error", err)
		d.mu.Lock()
		d.progress.Status = DurationRefreshFailed
		d.mu.Unlock()
		return
	}

	d.mu.Lock()
	d.progress.Total = len(songs)
	d.mu.Unlock()

	if len(songs) == 0 {
		return
	}

	for _, row := range songs {
		if ctx.Err() != nil {
			break
		}
		d.processOne(ctx, row)
	}
}

func (d *DurationRefresher) processOne(ctx context.Context, row sqlc.ListSongsNeedingDurationRow) {
	probeCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	url := row.Url
	if row.PluginEntryPath != "" && row.SourceData != "" {
		song := &models.Song{
			ID:              row.ID,
			PluginEntryPath: row.PluginEntryPath,
			SourceData:      row.SourceData,
			URL:             row.Url,
		}
		resolved, err := d.resolveURL(probeCtx, song)
		if err != nil {
			slog.Debug("duration refresh: resolve url failed", "songId", row.ID, "error", err)
			d.incFailed()
			return
		}
		url = resolved
	}

	if url == "" {
		d.incFailed()
		return
	}

	dur, err := d.extractor.ProbeDurationFromURL(probeCtx, url)
	if err != nil || dur <= 0 {
		slog.Debug("duration refresh: probe failed", "songId", row.ID, "error", err)
		d.incFailed()
		return
	}

	if err := d.updateDur(ctx, row.ID, dur); err != nil {
		slog.Warn("duration refresh: update failed", "songId", row.ID, "error", err)
		d.incFailed()
		return
	}

	d.mu.Lock()
	d.progress.Processed++
	d.mu.Unlock()
}

func (d *DurationRefresher) incFailed() {
	d.mu.Lock()
	d.progress.Failed++
	d.mu.Unlock()
}

var ErrDurationRefreshRunning = fmt.Errorf("duration refresh is already running")
