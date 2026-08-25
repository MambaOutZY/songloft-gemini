package jsplugin

import "testing"

// 覆盖 songloft-org/songloft#401 的权限门控：net:insecure-tls 必须显式声明。
//
// 关键点是它【不能】被 net 权限顺带覆盖——否则所有已声明 net 的插件（miot / bili /
// dav / pcyear-bridge …）会在升级后静默获得跳过证书校验的能力，而用户在权限列表里
// 看不到任何变化。

func TestPermNetInsecureTLS_NotImpliedByNet(t *testing.T) {
	// 只声明 net 的插件（现存插件的普遍形态）不得获得 insecure-tls
	if CheckPermission([]string{PermNet}, PermNetInsecureTLS) {
		t.Error("net 权限不应覆盖 net:insecure-tls——那会让现存插件静默获得跳过证书校验的能力")
	}
	// 反向也不成立：声明 insecure-tls 不等于拿到 net（原始 socket）
	if CheckPermission([]string{PermNetInsecureTLS}, PermNet) {
		t.Error("net:insecure-tls 不应覆盖 net")
	}
	// 显式声明才算有
	if !CheckPermission([]string{PermNet, PermNetInsecureTLS}, PermNetInsecureTLS) {
		t.Error("显式声明 net:insecure-tls 后应通过")
	}
}

// 现有的通配符糖不得成为绕过口子。
func TestPermNetInsecureTLS_NotGrantedByWildcards(t *testing.T) {
	for _, decl := range []string{"fs.*", "songs.*", "playlists.*"} {
		if CheckPermission([]string{decl}, PermNetInsecureTLS) {
			t.Errorf("通配符 %q 不应授予 net:insecure-tls", decl)
		}
	}
}

// 新权限必须进声明层白名单，否则 CheckPermission 认不出来（虽然
// ValidatePermissions 已改为宽容，白名单缺失仍会让权限彻底失效）。
func TestPermNetInsecureTLS_IsValidDeclaration(t *testing.T) {
	if err := ValidatePermissions([]string{PermNet, PermNetInsecureTLS}); err != nil {
		t.Fatalf("net:insecure-tls 应是合法声明: %v", err)
	}
}

// ValidatePermissions 对未知权限必须宽容：它挂在安装/更新的必经路径上，
// 拒绝会让声明了新权限的插件在旧宿主上直接装不了，砸掉插件其余正常功能。
// 宽容是安全中立的——CheckPermission 只认它认识的字符串。
func TestValidatePermissions_TolerantOfUnknown(t *testing.T) {
	if err := ValidatePermissions([]string{"storage", "net:insecure", "totally-made-up"}); err != nil {
		t.Fatalf("未知权限不应导致安装失败: %v", err)
	}
	// 宽容 ≠ 授权：拼错的权限名不得授予真权限
	if CheckPermission([]string{"net:insecure"}, PermNetInsecureTLS) {
		t.Error("拼错的 net:insecure 不得授予 net:insecure-tls")
	}
	if CheckPermission([]string{"totally-made-up"}, PermNetInsecureTLS) {
		t.Error("未知权限不得授予任何能力")
	}
}
