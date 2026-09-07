package services

import (
	"testing"
)

// items 按 filePath 构造待处理项，用于预建累积器的每目录待提取计数。
func items(paths ...string) []scanProcessItem {
	out := make([]scanProcessItem, 0, len(paths))
	for _, p := range paths {
		out = append(out, scanProcessItem{filePath: p})
	}
	return out
}

// batchPaths 提取批次里的文件路径，便于断言批次内容与顺序。
func batchPaths(batch []scanExtractResult) []string {
	out := make([]string, 0, len(batch))
	for _, r := range batch {
		out = append(out, r.item.filePath)
	}
	return out
}

// drain 反复取批直到取空，返回所有批次。
func drain(a *scanBatchAccumulator, size int, force bool) [][]string {
	var out [][]string
	for batch := a.takeBatch(size, force); batch != nil; batch = a.takeBatch(size, force) {
		out = append(out, batchPaths(batch))
	}
	return out
}

func TestScanBatchAccumulator_HoldsUntilDirComplete(t *testing.T) {
	all := items("/music/a/1.mp3", "/music/a/2.mp3", "/music/a/3.mp3")
	a := newScanBatchAccumulator(all, scanDirGroupCap)

	a.add(makeResult("/music/a/1.mp3", "T1", "Ar"))
	a.add(makeResult("/music/a/2.mp3", "T2", "Ar"))
	if got := a.takeBatch(1, true); got != nil {
		t.Fatalf("目录未提取完不应出批，却拿到 %v", batchPaths(got))
	}

	a.add(makeResult("/music/a/3.mp3", "T3", "Ar"))
	got := a.takeBatch(10, true)
	if len(got) != 3 {
		t.Fatalf("目录提取完应一次定稿 3 条，实际 %d 条", len(got))
	}
}

func TestScanBatchAccumulator_BatchSizeGating(t *testing.T) {
	// 三个目录各 1 个文件：每加一条就定稿一条，攒到 2 条才出批。
	all := items("/music/a/1.mp3", "/music/b/1.mp3", "/music/c/1.mp3")
	a := newScanBatchAccumulator(all, scanDirGroupCap)

	a.add(makeResult("/music/a/1.mp3", "T", "Ar"))
	if got := a.takeBatch(2, false); got != nil {
		t.Fatalf("不足批大小不应出批，却拿到 %v", batchPaths(got))
	}

	a.add(makeResult("/music/b/1.mp3", "T", "Ar"))
	got := a.takeBatch(2, false)
	if len(got) != 2 {
		t.Fatalf("攒够批大小应出 2 条，实际 %d 条", len(got))
	}
	if got := a.takeBatch(2, false); got != nil {
		t.Fatalf("队列已空不应再出批，却拿到 %v", batchPaths(got))
	}

	// 剩下那条只有 force 才拿得到。
	a.add(makeResult("/music/c/1.mp3", "T", "Ar"))
	if got := a.takeBatch(2, false); got != nil {
		t.Fatalf("残余不足批大小不应出批，却拿到 %v", batchPaths(got))
	}
	if got := a.takeBatch(2, true); len(got) != 1 {
		t.Fatalf("force 应取出残余 1 条，实际 %d 条", len(got))
	}
}

func TestScanBatchAccumulator_GroupCapSealsEarly(t *testing.T) {
	// 单目录 5 个文件、cap=2：不等目录扫完就应该分批定稿，
	// 这是「几千文件平铺在一个目录」不退化成全库累积的兜底。
	all := items("/music/flat/1.mp3", "/music/flat/2.mp3", "/music/flat/3.mp3",
		"/music/flat/4.mp3", "/music/flat/5.mp3")
	a := newScanBatchAccumulator(all, 2)

	a.add(makeResult("/music/flat/1.mp3", "T1", "Ar"))
	a.add(makeResult("/music/flat/2.mp3", "T2", "Ar"))
	got := a.takeBatch(10, true)
	if want := []string{"/music/flat/1.mp3", "/music/flat/2.mp3"}; len(got) != 2 {
		t.Fatalf("达到 groupCap 应提前定稿 %v，实际 %v", want, batchPaths(got))
	}

	a.add(makeResult("/music/flat/3.mp3", "T3", "Ar"))
	if got := a.takeBatch(10, true); len(got) != 0 {
		t.Fatalf("新一组未满 cap、目录也未扫完，不应出批，却拿到 %v", batchPaths(got))
	}
}

func TestScanBatchAccumulator_SealAllFlushesPartialGroups(t *testing.T) {
	// 模拟取消：目录 a 只提取了 1/3，pending 永远归不了零，
	// sealAll 必须把这一条冲出来，否则用户白干。
	all := items("/music/a/1.mp3", "/music/a/2.mp3", "/music/a/3.mp3", "/music/b/1.mp3")
	a := newScanBatchAccumulator(all, scanDirGroupCap)

	a.add(makeResult("/music/a/1.mp3", "T1", "Ar"))
	a.add(makeResult("/music/b/1.mp3", "T", "Ar"))
	if got := a.takeBatch(10, true); len(got) != 1 || batchPaths(got)[0] != "/music/b/1.mp3" {
		t.Fatalf("目录 b 已完整，应先出批，实际 %v", batchPaths(got))
	}

	a.sealAll()
	got := a.takeBatch(10, true)
	if len(got) != 1 || batchPaths(got)[0] != "/music/a/1.mp3" {
		t.Fatalf("sealAll 应冲出未完成目录的已提取结果，实际 %v", batchPaths(got))
	}
	if got := a.takeBatch(10, true); got != nil {
		t.Fatalf("sealAll 后不应残留，却拿到 %v", batchPaths(got))
	}
}

func TestScanBatchAccumulator_SealAllIdempotent(t *testing.T) {
	a := newScanBatchAccumulator(items("/music/a/1.mp3", "/music/a/2.mp3"), scanDirGroupCap)
	a.add(makeResult("/music/a/1.mp3", "T1", "Ar"))
	a.sealAll()
	a.sealAll()

	if got := a.takeBatch(10, true); len(got) != 1 {
		t.Fatalf("重复 sealAll 不应重复入队，实际 %d 条", len(got))
	}
	if got := a.takeBatch(10, true); got != nil {
		t.Fatalf("重复 sealAll 产生了重复结果 %v", batchPaths(got))
	}
}

func TestScanBatchAccumulator_AppliesFixSpamTagsPerGroup(t *testing.T) {
	// 定稿时必须已跑过 fixSpamTags：同目录 3 条垃圾 tag 回退到文件名，
	// 另一目录只有 1 条同样的 tag 则不受影响。
	all := items("/music/spam/a.mp3", "/music/spam/b.mp3", "/music/spam/c.mp3", "/music/ok/x.mp3")
	a := newScanBatchAccumulator(all, scanDirGroupCap)

	a.add(makeResult("/music/spam/a.mp3", "加微信", "广告"))
	a.add(makeResult("/music/spam/b.mp3", "加微信", "广告"))
	a.add(makeResult("/music/spam/c.mp3", "加微信", "广告"))
	spam := a.takeBatch(10, true)
	if len(spam) != 3 {
		t.Fatalf("垃圾目录应定稿 3 条，实际 %d 条", len(spam))
	}
	for _, r := range spam {
		if r.metadata.Title == "加微信" {
			t.Errorf("%s 的垃圾标题未回退到文件名", r.item.filePath)
		}
		if r.metadata.Artist != "" {
			t.Errorf("%s 的垃圾艺术家未清空，得到 %q", r.item.filePath, r.metadata.Artist)
		}
	}

	a.add(makeResult("/music/ok/x.mp3", "加微信", "广告"))
	ok := a.takeBatch(10, true)
	if len(ok) != 1 {
		t.Fatalf("正常目录应定稿 1 条，实际 %d 条", len(ok))
	}
	if ok[0].metadata.Title != "加微信" {
		t.Errorf("单条不构成垃圾 tag，标题不该被改，得到 %q", ok[0].metadata.Title)
	}
}

func TestScanBatchAccumulator_UnknownDirStillSeals(t *testing.T) {
	// 结果所在目录不在 pending 里（预过滤与提取之间目录被移动等边界情况）：
	// pending 为 0，应立即定稿而不是永久滞留。
	a := newScanBatchAccumulator(nil, scanDirGroupCap)
	a.add(makeResult("/music/ghost/1.mp3", "T", "Ar"))

	got := a.takeBatch(10, true)
	if len(got) != 1 {
		t.Fatalf("未登记目录的结果也应定稿，实际 %d 条", len(got))
	}
}

func TestScanBatchAccumulator_DrainsEverything(t *testing.T) {
	// 15 个文件散布在 5 个目录，批大小 4：所有结果都必须恰好出现一次。
	var paths []string
	for d := range 5 {
		for f := range 3 {
			paths = append(paths, string(rune('a'+d))+"/"+string(rune('1'+f))+".mp3")
		}
	}
	a := newScanBatchAccumulator(items(paths...), scanDirGroupCap)

	seen := make(map[string]int)
	for _, p := range paths {
		a.add(makeResult(p, "T", "Ar"))
		for _, batch := range drain(a, 4, false) {
			for _, got := range batch {
				seen[got]++
			}
		}
	}
	a.sealAll()
	for _, batch := range drain(a, 4, true) {
		for _, got := range batch {
			seen[got]++
		}
	}

	if len(seen) != len(paths) {
		t.Fatalf("出批文件数 = %d，want %d", len(seen), len(paths))
	}
	for _, p := range paths {
		if seen[p] != 1 {
			t.Errorf("%s 出批 %d 次，want 1", p, seen[p])
		}
	}
}
