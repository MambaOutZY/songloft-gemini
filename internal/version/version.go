package version

import "strings"

// 这些变量将在编译时通过 -ldflags 注入
var (
	Version   = "dev"
	GitCommit = "unknown"
	BuildTime = "unknown"
	BuildType = "" // 构建类型，"lite" 或 ""（默认 full）
)

// GetVersion 返回纯语义版本号（不含 v 前缀）
func GetVersion() string {
	return strings.TrimPrefix(Version, "v")
}

// GetFullVersion 返回包含 Git 提交和构建时间的完整版本信息
func GetFullVersion() string {
	s := GetVersion() + " (commit: " + GitCommit + ", built: " + BuildTime
	if BuildType != "" {
		s += ", type: " + BuildType
	}
	s += ")"
	return s
}
