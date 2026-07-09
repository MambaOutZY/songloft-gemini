package services

import (
	"slices"
	"strings"
	"testing"

	"songloft/pkg/cue"
)

// TestBuildSplitArgsFastSeek 验证 ffmpeg 切分参数使用「输入快速 seek」：
// -ss 必须出现在 -i 之前，否则对大 CD 镜像会退化成 O(N²) 全文件读取
// （songloft-org/songloft#260）。
func TestBuildSplitArgsFastSeek(t *testing.T) {
	track := cue.ResolvedTrack{
		AudioFilePath: "/music/CDImage.flac",
		StartSeconds:  3300,
		EndSeconds:    3600,
	}

	args := buildSplitArgs(track, "copy", "/out/track_15.flac")

	ssIdx := slices.Index(args, "-ss")
	iIdx := slices.Index(args, "-i")
	if ssIdx == -1 || iIdx == -1 {
		t.Fatalf("args missing -ss/-i: %v", args)
	}
	if ssIdx > iIdx {
		t.Errorf("-ss 必须在 -i 之前（输入快速 seek），got args=%v", args)
	}

	// 起点在 -ss 之后
	if args[ssIdx+1] != "3300.000" {
		t.Errorf("-ss value = %q, want 3300.000", args[ssIdx+1])
	}

	// 用 -t 时长裁剪而非 -to 绝对时间戳
	if slices.Contains(args, "-to") {
		t.Errorf("不应使用 -to（绝对时间戳），应使用 -t 时长: %v", args)
	}
	tIdx := slices.Index(args, "-t")
	if tIdx == -1 {
		t.Fatalf("缺少 -t 时长参数: %v", args)
	}
	if args[tIdx+1] != "300.000" {
		t.Errorf("-t value = %q, want 300.000 (EndSeconds-StartSeconds)", args[tIdx+1])
	}
}

// TestBuildSplitArgsLastTrackNoDuration 最后一个 track（EndSeconds 为 0）
// 不带 -t，直接读到文件末尾。
func TestBuildSplitArgsLastTrackNoDuration(t *testing.T) {
	track := cue.ResolvedTrack{
		AudioFilePath: "/music/CDImage.flac",
		StartSeconds:  3600,
		EndSeconds:    0,
	}

	args := buildSplitArgs(track, "copy", "/out/track_20.flac")

	if slices.Contains(args, "-t") {
		t.Errorf("EndSeconds=0 时不应带 -t: %v", args)
	}
	// 仍应快速 seek 到起点
	ssIdx := slices.Index(args, "-ss")
	iIdx := slices.Index(args, "-i")
	if ssIdx == -1 || iIdx == -1 || ssIdx > iIdx {
		t.Errorf("-ss 应在 -i 之前: %v", args)
	}
	// 结尾仍应有输出参数
	if !strings.HasSuffix(args[len(args)-1], "track_20.flac") {
		t.Errorf("最后一个参数应为输出路径: %v", args)
	}
}
