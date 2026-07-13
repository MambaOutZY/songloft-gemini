package services

import "sync/atomic"

// DownloadActivity 跟踪进行中的歌曲下载数量，供后台任务判断是否应为下载让路。
//
// 背景（issue #265）：导入 100 首歌单时 probeRemoteSongsMetadata 会并发探测（每个探测
// 走 ytdlp 插件 + ffprobe 拉流），CPU 打满；紧接着的批量下载解析 music/url 走同一个
// ytdlp 插件唯一 worker，排在探测 backlog 后面超过 30s 被判死。让探测在有活跃下载时退避，
// 从源头避免争用。
//
// 零值可用；方法并发安全。
type DownloadActivity struct {
	n atomic.Int64
}

// Begin 标记一个下载开始（进入临界区）。
func (a *DownloadActivity) Begin() {
	if a == nil {
		return
	}
	a.n.Add(1)
}

// End 标记一个下载结束。与 Begin 成对调用（通常 defer）。
func (a *DownloadActivity) End() {
	if a == nil {
		return
	}
	a.n.Add(-1)
}

// Active 是否有进行中的下载。
func (a *DownloadActivity) Active() bool {
	if a == nil {
		return false
	}
	return a.n.Load() > 0
}
