package services

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"songloft/internal/database"
	"songloft/internal/models"
)

// SongTagRepository 是 SongTagService 依赖的标签仓储接口。
type SongTagRepository interface {
	Create(ctx context.Context, name, color string) (int64, error)
	GetByID(ctx context.Context, id int64) (*models.SongTag, error)
	GetByName(ctx context.Context, name string) (*models.SongTag, error)
	Update(ctx context.Context, id int64, name, color string) error
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context, keyword, orderBy, order string, limit, offset int) ([]models.SongTag, error)
	Count(ctx context.Context, keyword string) (int64, error)
	GetBySongID(ctx context.Context, songID int64) ([]models.SongTag, error)
	SetSongTags(ctx context.Context, songID int64, tagIDs []int64) error
	LinkSongTag(ctx context.Context, songID, tagID int64) error
	UnlinkSongTag(ctx context.Context, songID, tagID int64) error
	ListSongIDs(ctx context.Context, tagID int64, limit, offset int) ([]int64, error)
	CountSongs(ctx context.Context, tagID int64) (int64, error)
}

// SongTagService 自定义标签服务
type SongTagService struct {
	tags SongTagRepository
}

func NewSongTagService(tags SongTagRepository) *SongTagService {
	return &SongTagService{tags: tags}
}

func (s *SongTagService) Create(ctx context.Context, name, color string) (*models.SongTag, error) {
	name = strings.TrimSpace(name)
	if err := validateTagName(name); err != nil {
		return nil, err
	}
	id, err := s.tags.Create(ctx, name, color)
	if err != nil {
		return nil, fmt.Errorf("song_tag_service: create: %w", err)
	}
	return s.tags.GetByID(ctx, id)
}

func (s *SongTagService) GetByID(ctx context.Context, id int64) (*models.SongTag, error) {
	return s.tags.GetByID(ctx, id)
}

func (s *SongTagService) Update(ctx context.Context, id int64, name, color string) error {
	name = strings.TrimSpace(name)
	if err := validateTagName(name); err != nil {
		return err
	}
	tag, err := s.tags.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if name == "" {
		name = tag.Name
	}
	if color == "" {
		color = tag.Color
	}
	return s.tags.Update(ctx, id, name, color)
}

func (s *SongTagService) Delete(ctx context.Context, id int64) error {
	return s.tags.Delete(ctx, id)
}

func (s *SongTagService) List(ctx context.Context, keyword, orderBy, order string, limit, offset int) ([]models.SongTag, error) {
	return s.tags.List(ctx, keyword, orderBy, order, limit, offset)
}

func (s *SongTagService) Count(ctx context.Context, keyword string) (int64, error) {
	return s.tags.Count(ctx, keyword)
}

func (s *SongTagService) GetSongTags(ctx context.Context, songID int64) ([]models.SongTag, error) {
	return s.tags.GetBySongID(ctx, songID)
}

func (s *SongTagService) SetSongTags(ctx context.Context, songID int64, tagIDs []int64) error {
	return s.tags.SetSongTags(ctx, songID, tagIDs)
}

func (s *SongTagService) BatchBind(ctx context.Context, tagID int64, songIDs []int64) (int, error) {
	bound := 0
	for _, songID := range songIDs {
		if err := s.tags.LinkSongTag(ctx, songID, tagID); err != nil {
			return bound, fmt.Errorf("song_tag_service: bind song %d: %w", songID, err)
		}
		bound++
	}
	return bound, nil
}

func (s *SongTagService) BatchUnbind(ctx context.Context, tagID int64, songIDs []int64) (int, error) {
	unbound := 0
	for _, songID := range songIDs {
		if err := s.tags.UnlinkSongTag(ctx, songID, tagID); err != nil {
			return unbound, fmt.Errorf("song_tag_service: unbind song %d: %w", songID, err)
		}
		unbound++
	}
	return unbound, nil
}

func (s *SongTagService) ListSongIDs(ctx context.Context, tagID int64, limit, offset int) ([]int64, error) {
	return s.tags.ListSongIDs(ctx, tagID, limit, offset)
}

func (s *SongTagService) CountSongs(ctx context.Context, tagID int64) (int64, error) {
	return s.tags.CountSongs(ctx, tagID)
}

// FromPlaylist 将歌单内所有歌曲关联到同名标签（标签不存在则创建）。
func (s *SongTagService) FromPlaylist(ctx context.Context, playlistName string, songIDs []int64) (*models.SongTag, int, error) {
	name := strings.TrimSpace(playlistName)
	if err := validateTagName(name); err != nil {
		return nil, 0, err
	}

	tag, err := s.tags.GetByName(ctx, name)
	if err != nil {
		if !errors.Is(err, database.ErrNotFound) {
			return nil, 0, err
		}
		id, createErr := s.tags.Create(ctx, name, "")
		if createErr != nil {
			return nil, 0, fmt.Errorf("song_tag_service: from_playlist create: %w", createErr)
		}
		tag, err = s.tags.GetByID(ctx, id)
		if err != nil {
			return nil, 0, err
		}
	}

	bound := 0
	for _, songID := range songIDs {
		if err := s.tags.LinkSongTag(ctx, songID, tag.ID); err != nil {
			return tag, bound, err
		}
		bound++
	}
	return tag, bound, nil
}

func validateTagName(name string) error {
	if name == "" {
		return fmt.Errorf("标签名不能为空")
	}
	if len([]rune(name)) > 50 {
		return fmt.Errorf("标签名不能超过 50 个字符")
	}
	if strings.Contains(name, ",") {
		return fmt.Errorf("标签名不能包含逗号")
	}
	return nil
}
