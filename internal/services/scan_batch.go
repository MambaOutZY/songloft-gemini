package services

import "path/filepath"

// scanDirGroupCap 单个目录组累积的结果上限。
// 兜住「几千个文件平铺在同一个目录」的极端曲库：达到上限即提前定稿出批，
// 避免退化成全库累积。
const scanDirGroupCap = 200

// scanBatchAccumulator 按目录分组累积元数据提取结果，供流式入库使用。
//
// 为什么不能直接「攒满 dbBatchSize 就入库」：fixSpamTags 按目录分组、以
// 「同目录内超过一半文件共享同一 (title, artist)」判定垃圾 tag，需要整个目录
// 作为样本。所以这里以目录为定稿单位——某目录的全部文件都提取完成（或该组累积
// 达到 groupCap）时，才对该组跑 fixSpamTags 并转入待入库队列。
//
// 收益：内存占用从 O(全库) 收敛到 O(groupCap + 待入库队列)，且第一批结果能在
// 秒级落库，前端进度与曲库随即可见（songloft-org/songloft#447）。
type scanBatchAccumulator struct {
	groupCap int

	// pending 记录每个目录还有多少文件尚未提取完成。
	pending map[string]int
	// groups 记录每个目录已提取、尚未定稿的结果。
	groups map[string][]scanExtractResult
	// ready 已定稿（已跑过 fixSpamTags）、等待入库的结果，保持定稿顺序。
	ready []scanExtractResult
}

// newScanBatchAccumulator 按待处理项预建每个目录的待提取计数。
func newScanBatchAccumulator(items []scanProcessItem, groupCap int) *scanBatchAccumulator {
	if groupCap <= 0 {
		groupCap = scanDirGroupCap
	}
	a := &scanBatchAccumulator{
		groupCap: groupCap,
		pending:  make(map[string]int),
		groups:   make(map[string][]scanExtractResult),
	}
	for _, item := range items {
		a.pending[filepath.Dir(item.filePath)]++
	}
	return a
}

// add 收下一条提取结果；该结果所在目录提取完毕或组内累积达到上限时定稿。
func (a *scanBatchAccumulator) add(r scanExtractResult) {
	dir := filepath.Dir(r.item.filePath)
	a.groups[dir] = append(a.groups[dir], r)

	if n := a.pending[dir]; n > 0 {
		a.pending[dir] = n - 1
	}

	if a.pending[dir] == 0 || len(a.groups[dir]) >= a.groupCap {
		a.seal(dir)
	}
}

// sealAll 定稿所有仍未提取完的目录组，用于取消或收尾。
// worker 提前退出（取消）时 pending 永远归不了零，靠这里把已提取的成果冲出来。
func (a *scanBatchAccumulator) sealAll() {
	for dir := range a.groups {
		a.seal(dir)
	}
}

// seal 对单个目录组跑 fixSpamTags 后转入待入库队列。
// fixSpamTags 内部自行按目录分组，传单目录切片即恒等包装。
func (a *scanBatchAccumulator) seal(dir string) {
	group := a.groups[dir]
	if len(group) == 0 {
		delete(a.groups, dir)
		return
	}
	fixSpamTags(group)
	a.ready = append(a.ready, group...)
	// 目录内仍有文件未提取时（达到 groupCap 提前定稿），后续结果会重开一组。
	delete(a.groups, dir)
}

// takeBatch 取出一个待入库批次。
// force 为 false 时不足 size 返回 nil（等更多结果定稿再凑批）；
// force 为 true 时返回剩余的全部（不足 size 也返回），队列空则返回 nil。
func (a *scanBatchAccumulator) takeBatch(size int, force bool) []scanExtractResult {
	if size <= 0 {
		size = dbBatchSize
	}
	if len(a.ready) == 0 {
		return nil
	}
	if len(a.ready) < size && !force {
		return nil
	}
	n := min(size, len(a.ready))
	batch := a.ready[:n]
	a.ready = a.ready[n:]
	return batch
}
