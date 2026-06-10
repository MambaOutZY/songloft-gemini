// Package tracelycfg holds Tracely client configuration injected at build time.
//
// AppID、AppSecret 与 Host 通过 -ldflags "-X songloft/internal/tracelycfg.XXX=..." 注入，
// 仅在私有构建时由 CI/Makefile 提供。开源构建默认三者为空，对应 Tracely 客户端不会被初始化。
package tracelycfg

var (
	AppID     = ""
	AppSecret = ""
	Host      = ""
)

// Enabled 报告是否启用 Tracely 上报（AppID、AppSecret 与 Host 都已注入）。
func Enabled() bool {
	return AppID != "" && AppSecret != "" && Host != ""
}
