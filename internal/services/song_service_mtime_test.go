package services

import (
	"testing"
	"time"
)

// TestFileModTimeMatches 覆盖扫描跳过条件对 file_modified_at 的判断：
// 用户修改内嵌 USLT 后磁盘 mtime 会变化，若旧 mtime 与新 mtime 一致才允许跳过扫描；
// 存量老记录 stored=nil 时保持兼容，视为匹配（不会触发全库重扫）。
func TestFileModTimeMatches(t *testing.T) {
	base := time.Date(2026, 9, 21, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name    string
		stored  *time.Time
		current time.Time
		want    bool
	}{
		{
			name:    "nil stored treats as match (legacy row)",
			stored:  nil,
			current: base,
			want:    true,
		},
		{
			name:    "equal to the second",
			stored:  timePtr(base),
			current: base.Add(500 * time.Millisecond),
			want:    true,
		},
		{
			name:    "differs by one second",
			stored:  timePtr(base),
			current: base.Add(time.Second),
			want:    false,
		},
		{
			name:    "current earlier than stored",
			stored:  timePtr(base),
			current: base.Add(-2 * time.Second),
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := fileModTimeMatches(tt.stored, tt.current)
			if got != tt.want {
				t.Fatalf("fileModTimeMatches(%v, %v) = %v, want %v", tt.stored, tt.current, got, tt.want)
			}
		})
	}
}

func timePtr(t time.Time) *time.Time { return &t }
