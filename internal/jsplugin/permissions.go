package jsplugin

import (
	"log/slog"
	"strings"
)

// Permission 权限常量
//
// 执行层（runtime）权限 = 插件可申明的细粒度权限。
// 声明层允许题外子集： songs.* / playlists.* 仅作为“一把梭”通配符糖，
// 以配合 CheckPermission 的前缀匹配实现。
const (
	PermStorage           = "storage"            // 持久化存储
	PermSongsRead         = "songs.read"         // 读取歌曲
	PermSongsWrite        = "songs.write"        // 修改歌曲
	PermPlaylistsRead     = "playlists.read"     // 读取歌单
	PermPlaylistsWrite    = "playlists.write"    // 修改歌单
	PermInterPlugin       = "inter-plugin"       // 插件间通信
	PermCommand           = "command"            // 执行命令
	PermJSEnv             = "jsenv"              // 创建/执行子 JS 环境（songloft.jsenv.*）
	PermFS                = "fs"                 // 插件数据目录内文件读写
	PermFSMusic           = "fs:music"           // 可访问 music_path 音乐目录
	PermFSExternal        = "fs:external"        // 可访问管理员配置的外部目录
	PermWebSocket         = "websocket"          // WebSocket 连接
	PermPersistentStorage = "persistent-storage" // 持久化存储（卸载后保留）
	PermNet               = "net"                // 原始网络 socket（UDP）
	// PermNetInsecureTLS 允许 fetch 带 X-Fetch-Insecure 跳过 TLS 证书校验。
	// 独立于 PermNet：CheckPermission 无前缀匹配（"net" 不覆盖 "net:insecure-tls"），
	// 故必须在 manifest 里显式声明，用户装插件时能在权限列表看到。
	// 用途：自建 NAS 设备（飞牛 fnOS 5667 等）默认自签证书、且插件按裸 IP 访问，
	// 证书结构性无法校验（songloft-org/songloft#401）。
	PermNetInsecureTLS = "net:insecure-tls"
)

// AllPermissions 所有合法权限列表（声明层白名单）。
// 包含两个通配符糖 songs.* / playlists.*：仅当作声明时的便捷写法，
// 实际走 CheckPermission 的前缀匹配，runtime 层的 action 仍会被
// extractPermFromAction 映射为细粒度权限。
var AllPermissions = []string{
	PermStorage,
	PermSongsRead, PermSongsWrite,
	PermPlaylistsRead, PermPlaylistsWrite,
	PermInterPlugin, PermCommand,
	PermJSEnv, PermFS, PermFSMusic, PermFSExternal,
	PermWebSocket, PermPersistentStorage, PermNet, PermNetInsecureTLS,
	// 通配符糖
	"songs.*",
	"playlists.*",
	"fs.*",
}

// CheckPermission 检查插件是否拥有指定权限
// 支持通配符匹配：如 "playlists.*" 匹配 "playlists.read"
func CheckPermission(permissions []string, required string) bool {
	for _, p := range permissions {
		if p == required {
			return true
		}
		// 通配符匹配：如果权限以 ".*" 结尾，则匹配相同前缀的所有子权限
		if strings.HasSuffix(p, ".*") {
			prefix := strings.TrimSuffix(p, ".*")
			if strings.HasPrefix(required, prefix+".") || required == prefix {
				return true
			}
		}
	}
	return false
}

// ValidatePermissions 检查权限声明，未知权限只告警不拒绝。
//
// 为什么不 return error：本函数挂在安装（InstallFromUpload）与更新（Update）的
// 必经路径上，一旦拒绝，插件在【旧宿主上直接装不了】。于是每新增一个权限常量，
// 都会让声明了它的新版插件在所有存量宿主上安装失败——报的还是 `unknown
// permission: "..."` 这种用户完全无法自救的错，且砸掉的是插件原本正常的其他功能。
// songloft-org/songloft#401 新增 net:insecure-tls 时踩到这一点。
//
// 宽容是安全中立的：CheckPermission 只认它认识的字符串，未知权限授不了任何能力。
// 代价是权限名拼错（如 net:insecure 漏了 -tls）不再是硬失败，故必须留下 Warn 日志
// 让作者能查到。
//
// 返回类型保留 error（恒为 nil）以免动所有调用点的签名。
func ValidatePermissions(permissions []string) error {
	valid := make(map[string]bool, len(AllPermissions))
	for _, p := range AllPermissions {
		valid[p] = true
	}
	for _, p := range permissions {
		if !valid[p] {
			slog.Warn("插件声明了未知权限，已忽略（可能是权限名拼写错误，或该权限需要更高版本宿主）",
				"permission", p)
		}
	}
	return nil
}
