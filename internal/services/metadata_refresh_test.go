package services

import "testing"

// TestDropFilenameFallbackTitle 验证缓存回填时对「用文件名当 title」回退结果的拦截：
// 缓存文件名形如 `{songID}.{plugin_entry_path}_{dedup_key}`，绝不是真实歌名，
// 若 Extract 因流无内嵌标题而回退成该文件名，必须清空，避免写进 songs.title
// （songloft-org/songloft#286）。真实 tag 标题（不等于文件名 base）应原样保留。
func TestDropFilenameFallbackTitle(t *testing.T) {
	cases := []struct {
		name     string
		title    string
		filePath string
		want     string
	}{
		{
			name:     "缓存文件名回退 → 清空",
			title:    "3785.ytdlp_youtube_qSnSCOMVSYQ_p49",
			filePath: "/data/music_cache/37/0/3785.ytdlp_youtube_qSnSCOMVSYQ_p49.mp3",
			want:     "",
		},
		{
			name:     "真实歌名 → 保留",
			title:    "林憶蓮 Sandy Lam【聽說愛情回來過】",
			filePath: "/data/music_cache/37/0/3785.ytdlp_youtube_qSnSCOMVSYQ_p49.mp3",
			want:     "林憶蓮 Sandy Lam【聽說愛情回來過】",
		},
		{
			name:     "空标题 → 保持空",
			title:    "",
			filePath: "/data/music_cache/37/0/3785.ytdlp_youtube_qSnSCOMVSYQ_p49.mp3",
			want:     "",
		},
		{
			// Extract 的 fileName = TrimSuffix(base, Ext)，本 helper 用完全相同的逻辑，
			// 二者对同一路径必然一致：此处 base="3785.ytdlp_youtube_qSnSCOMVSYQ_p49"、Ext=".opus"。
			name:     "opus 缓存文件名回退 → 清空",
			title:    "3785.ytdlp_youtube_qSnSCOMVSYQ_p49",
			filePath: "/data/music_cache/37/0/3785.ytdlp_youtube_qSnSCOMVSYQ_p49.opus",
			want:     "",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := dropFilenameFallbackTitle(tc.title, tc.filePath); got != tc.want {
				t.Errorf("dropFilenameFallbackTitle(%q, %q) = %q, want %q", tc.title, tc.filePath, got, tc.want)
			}
		})
	}
}
