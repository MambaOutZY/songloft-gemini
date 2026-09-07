package logging

import (
	"bufio"
	"io"
	"regexp"
	"runtime"
	"strings"
	"sync"
)

// 脱敏规则集合。导出日志给用户提交 issue 前，逐条 apply 到日志文本，抹掉：
//   - 密钥/token/密码等键值（jwt_secret=xxx、"access_token":"xxx"）
//   - Authorization / Cookie 头
//   - Bearer token
//   - URL 里的内嵌凭证（scheme://user:pass@host）
//   - 客户端 IP（保留网段，抹掉主机位）
//   - 用户主目录路径中的用户名段（/home/<user>、/Users/<user>、C:\Users\<user>）
//
// 这是「导出时」的防御纵深：写入侧已尽量不打印明文密码/凭证（见 app.go、runtime.go），
// 但插件对外请求头等仍可能在 debug 级别落日志，故导出前统一再过一遍。
var (
	// 键值型密钥：key<sep>value。sep 覆盖 slog 文本格式（key=value）与 JSON（"key":"value"）。
	// value 取到下一个分隔符（空白、引号、逗号、右括号）为止。
	reSecretKV = regexp.MustCompile(`(?i)(jwt_secret|refresh_token|access_token|api_?key|secret|password|passwd|token)(["']?\s*[:=]\s*["']?)([^"'\s,}\]]+)`)

	// 中文标签密钥：默认密码: xxx / 密钥：xxx（防御纵深，兼顾中文日志）。
	// 要求标签后带分隔符或空格，避免误伤「密钥已生成」这类无值的正常日志。
	reSecretCN = regexp.MustCompile(`(密码|密钥|令牌)([:：=]\s*|\s+)(\S+)`)

	// 认证类头部（map/JSON 形式）：Authorization:xxx、Cookie:xxx、Set-Cookie:xxx。
	reAuthHeader = regexp.MustCompile(`(?i)(authorization|set-cookie|cookie)(["']?\s*[:=]\s*["']?)([^"'\s,}\]]+)`)

	// Bearer token（值中带空格，单独处理）。
	reBearer = regexp.MustCompile(`(?i)bearer\s+[A-Za-z0-9._~+/=-]+`)

	// URL 内嵌凭证：scheme://user:pass@host。
	reURLCred = regexp.MustCompile(`([a-zA-Z][a-zA-Z0-9+.\-]*://)([^:/@\s]+):([^@/\s]+)@`)

	// IPv4 字面量。用 ReplaceAllFunc 保留网段、跳过回环/未指定地址。
	reIPv4 = regexp.MustCompile(`\b\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}\b`)

	// Unix 用户主目录：/home/<user> 或 /Users/<user>。
	reUnixHome = regexp.MustCompile(`(/home/|/Users/)([^/\s]+)`)

	// Windows 用户目录：C:\Users\<user>。
	reWinHome = regexp.MustCompile(`(?i)([A-Z]:\\Users\\)([^\\/\s]+)`)
)

// Redact 返回脱敏后的文本。对整段文本一次性 apply 所有规则，可安全用于多行内容。
//
// 四条规则前面挂了字面量闸门。这不是微优化：导出端点要过 10 MiB，逐条实测
// （21740 行真实日志）里 reWinHome 147ms、reURLCred 252ms、reIPv4 141ms，而它们
// 在一份 Linux/macOS 服务器日志里几乎不可能命中——`strings.Contains` 走 memchr，
// 比让 RE2 把整行扫一遍便宜一到两个数量级。
//
// ⚠️ 闸门只加在**不依赖 `(?i)` 字母匹配**的规则上，这是硬约束不是偏好。
// Go 的 `(?i)` 做的是 Unicode 简单折叠：`(?i)secret` 也匹配 `ſecret`（U+017F）、
// `(?i)token` 也匹配带 Kelvin 记号（U+212A）的 `toKen`。用 ASCII 折叠去闸门
// reSecretKV / reAuthHeader / reBearer，在真实日志上测不出差异，但那是经验一致
// 而非可证等价——这是脱敏路径，漏一行就是把密钥贴到公开 issue 上，不值得下这个注。
// 下面四条闸门的判据全是非字母字面量（`://`、`:\`、`/home/`、`.`）或非折叠字符
// （密/令），大小写折叠对它们没有作用域，因此等价可证。
func Redact(s string) string {
	s = reSecretKV.ReplaceAllString(s, "${1}${2}***")
	if strings.ContainsAny(s, "密令") { // 无 (?i)，中日韩字符不参与大小写折叠
		s = reSecretCN.ReplaceAllString(s, "${1}${2}***")
	}
	// 先处理 Bearer（值带空格），再处理认证头，避免认证头规则把 "Bearer <token>" 截断成半个。
	s = reBearer.ReplaceAllString(s, "Bearer ***")
	s = reAuthHeader.ReplaceAllStringFunc(s, func(m string) string {
		sub := reAuthHeader.FindStringSubmatch(m)
		if len(sub) < 4 {
			return m
		}
		if strings.EqualFold(sub[3], "Bearer") { // 已由 reBearer 处理，勿再截断
			return m
		}
		return sub[1] + sub[2] + "***"
	})
	if strings.Contains(s, "://") { // 无 (?i)，且 scheme 分隔符是模式里的必要字面量
		s = reURLCred.ReplaceAllString(s, "${1}***:***@")
	}
	if strings.ContainsRune(s, '.') { // 无 (?i)，IPv4 字面量必然含点
		s = reIPv4.ReplaceAllStringFunc(s, redactIPv4)
	}
	if strings.Contains(s, "/home/") || strings.Contains(s, "/Users/") { // 无 (?i)
		s = reUnixHome.ReplaceAllString(s, "${1}<user>")
	}
	// reWinHome 带 (?i)，但闸门判据 `:\` 只由非字母组成，不受折叠影响。
	if strings.Contains(s, `:\`) {
		s = reWinHome.ReplaceAllString(s, "${1}<user>")
	}
	return s
}

// redactIPv4 保留前两段网段、抹掉主机位（1.2.3.4 → 1.2.*.*）。
// 回环（127.x）与未指定地址（0.0.0.0）不脱敏——它们不含隐私且对排查本地/绑定问题有用。
func redactIPv4(ip string) string {
	parts := strings.Split(ip, ".")
	if len(parts) != 4 {
		return ip
	}
	if parts[0] == "127" || ip == "0.0.0.0" {
		return ip
	}
	return parts[0] + "." + parts[1] + ".*.*"
}

const (
	// 单行上限：bufio.Scanner 默认 64KiB，放宽到 1MiB 以容纳偶发的长行。
	redactMaxLineBytes = 1 << 20
	// 每个并行块的大小上限，字节与行数两条一起卡，任一条满就派活。
	//
	// 以**字节**为主：RedactStream 的既有约束是「不把整份日志读进内存」，而单行
	// 上限是 1 MiB，只按行数切块的话 512 行最坏能攒到 512 MB，等于把那条约束丢了。
	// 按 64 KiB 切，在飞内存就与行长无关：块数被通道容量卡在 3×workers 以内，
	// in 与 out 各一份，典型也就 3 MB 上下。
	//
	// 行数上限仍保留：对极短行（比如整片空行）避免一个块攒进上万行、
	// 让块粒度大到影响并行度。
	redactChunkBytes    = 64 * 1024
	redactChunkMaxLines = 512
	// 并行上限。实测 12 核机器上 ×8 已拿到 5.9x、再往上基本持平，
	// 而导出是低频操作，不该为它占满整机 CPU。
	redactMaxWorkers = 8
	// 写出缓冲。此前每行两次 io.WriteString 直捅 http.ResponseWriter，
	// 10 MiB 是 4 万多次小写；聚成 64KiB 一次交给下游（含 gzip 中间件）。
	redactWriteBufBytes = 64 * 1024
)

// RedactStream 逐行读取 src、脱敏后写入 dst，避免把整份日志读进内存。
//
// 为什么要并行：脱敏本身就是这个端点的全部成本。实测 10 MiB（21740 行）里
// RedactStream 花 1.889s，而纯逐行扫描只花 0.59ms——99.97% 是 8 遍正则的 CPU。
// 客户端（Lynx 与 Flutter 都一样）在这 1.9s 里只能干等分享面板，正是
// songloft-player-lynx#3 在把打包下移到原生之后剩下的那一段。
//
// 并行按「块内逐行调同一个 Redact、块间严格保序」实现，所以输出与串行逐字节一致——
// 这是按构造成立的，不依赖对日志内容的假设。GOMAXPROCS 为 1 时退回串行，
// 避免为单核机器付调度开销。
func RedactStream(dst io.Writer, src io.Reader) error {
	workers := runtime.GOMAXPROCS(0)
	if workers > redactMaxWorkers {
		workers = redactMaxWorkers
	}
	if workers < 2 {
		return redactStreamSerial(dst, src)
	}
	return redactStreamParallel(dst, src, workers)
}

func redactStreamSerial(dst io.Writer, src io.Reader) error {
	bw := bufio.NewWriterSize(dst, redactWriteBufBytes)
	sc := bufio.NewScanner(src)
	sc.Buffer(make([]byte, 0, 64*1024), redactMaxLineBytes)
	for sc.Scan() {
		if _, err := bw.WriteString(Redact(sc.Text())); err != nil {
			return err
		}
		if err := bw.WriteByte('\n'); err != nil {
			return err
		}
	}
	if err := sc.Err(); err != nil {
		return err
	}
	return bw.Flush()
}

// redactBlockFull 判断当前攒着的行是否够派一个块了。
//
// 抽成函数是为了让「字节上限」这条真的能被测到：它只影响内存占用，不影响输出，
// 所以把 `bytes >= redactChunkBytes` 删掉的话，任何比对输出的用例都是绿的，
// 而 512 行 × 1 MiB 单行上限的最坏情况会悄悄回来。
func redactBlockFull(lines, bytes int) bool {
	return bytes >= redactChunkBytes || lines >= redactChunkMaxLines
}

// redactBlock 一个待脱敏的行块。ready 关闭即表示 out 已就绪。
type redactBlock struct {
	in    []string
	out   []string
	ready chan struct{}
}

func redactStreamParallel(dst io.Writer, src io.Reader, workers int) error {
	// work 派活，order 定序。两条通道都有界，生产者被下游的消费速度反压，
	// 所以在飞的块数不超过 3×workers。
	work := make(chan *redactBlock, workers)
	order := make(chan *redactBlock, workers*2)
	// 下游写失败（最典型是客户端断开）时通知生产者收工，别把整份日志读完。
	abort := make(chan struct{})

	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for block := range work {
				block.out = make([]string, len(block.in))
				for i, line := range block.in {
					block.out[i] = Redact(line)
				}
				close(block.ready)
			}
		}()
	}

	var scanErr error
	go func() {
		// 先派活再入队，顺序不能反：消费者会在 order 里取到的块上等 ready，
		// 若某块入了 order 却因 abort 没派出去，那个 ready 永远不会关，消费者就吊死。
		defer close(order)
		defer close(work)

		sc := bufio.NewScanner(src)
		sc.Buffer(make([]byte, 0, 64*1024), redactMaxLineBytes)
		lines := make([]string, 0, 128)
		pending := 0

		dispatch := func() bool {
			if len(lines) == 0 {
				return true
			}
			block := &redactBlock{in: lines, ready: make(chan struct{})}
			lines = make([]string, 0, 128)
			pending = 0
			select {
			case work <- block:
			case <-abort:
				return false
			}
			select {
			case order <- block:
			case <-abort:
				return false
			}
			return true
		}

		for sc.Scan() {
			// sc.Text() 复制出字符串：sc.Bytes() 的底层数组下一次 Scan 就会被覆写，
			// 而这些行要交给别的 goroutine 处理。
			line := sc.Text()
			lines = append(lines, line)
			pending += len(line)
			if redactBlockFull(len(lines), pending) {
				if !dispatch() {
					return
				}
			}
		}
		if !dispatch() {
			return
		}
		// 在 close(order) 之前赋值（defer 后于函数体执行），消费者观察到通道关闭
		// 即 happens-after 这次写入，无需额外同步。
		scanErr = sc.Err()
	}()

	bw := bufio.NewWriterSize(dst, redactWriteBufBytes)
	var writeErr error
	// 第一次失败才关 abort。abort 关两次会 panic，而这里是请求处理路径，
	// 所以把「只关一次」做成显式不变式，不靠读者去追下面的 continue 分支。
	fail := func(err error) {
		if writeErr == nil {
			writeErr = err
			close(abort)
		}
	}
	for block := range order {
		<-block.ready
		if writeErr != nil {
			continue // 仍要排空 order，否则生产者卡在 order <- block 上
		}
		for _, line := range block.out {
			if _, err := bw.WriteString(line); err != nil {
				fail(err)
				break
			}
			if err := bw.WriteByte('\n'); err != nil {
				fail(err)
				break
			}
		}
	}
	wg.Wait()

	if writeErr != nil {
		return writeErr
	}
	if scanErr != nil {
		return scanErr
	}
	return bw.Flush()
}
