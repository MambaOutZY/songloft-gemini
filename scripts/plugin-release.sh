#!/bin/bash
# 触发插件的 GitHub Action Release workflow
# 用法:
#   ./scripts/plugin-release.sh <plugin-name> [version]
#   ./scripts/plugin-release.sh all [version]
#
# 示例:
#   ./scripts/plugin-release.sh miot              # 发布 miot 插件（自动生成日期版本）
#   ./scripts/plugin-release.sh miot 2026.4.20    # 发布 miot 插件（指定版本）
#   ./scripts/plugin-release.sh all               # 发布所有插件
#   ./scripts/plugin-release.sh all 2026.4.20     # 发布所有插件（指定版本）

set -e

# 脚本所在目录 / 仓库根目录 / 插件源目录
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
PLUGINS_SRC_DIR="${REPO_ROOT}/jsplugins-src"

# 插件列表：扫描 jsplugins-src/songloft-plugin-* 目录，去掉前缀得到插件名
PLUGINS=()
for dir in "${PLUGINS_SRC_DIR}"/songloft-plugin-*/; do
    [ -d "$dir" ] || continue
    name="$(basename "$dir")"
    PLUGINS+=("${name#songloft-plugin-}")
done

# 颜色输出
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

usage() {
    echo "用法: $0 <plugin-name|all> [version]"
    echo ""
    echo "可用插件:"
    for p in "${PLUGINS[@]}"; do
        echo "  - $p"
    done
    echo "  - all (所有插件)"
    echo ""
    echo "示例:"
    echo "  $0 miot              # 自动生成日期版本"
    echo "  $0 miot 2026.4.20    # 指定版本号"
    echo "  $0 all               # 发布所有插件"
    exit 1
}

# 触发单个插件的 Release workflow
release_plugin() {
    local plugin_name="$1"
    local version="$2"
    local plugin_dir="${PLUGINS_SRC_DIR}/songloft-plugin-${plugin_name}"

    # 从子目录 git remote 解析 owner/repo（兼容 git@github.com: 和 https:// 两种格式）
    local remote_url
    remote_url="$(git -C "${plugin_dir}" remote get-url origin 2>/dev/null)"
    if [ -z "$remote_url" ]; then
        echo -e "${RED}✗ ${plugin_name} 无法读取 git remote（${plugin_dir}）${NC}"
        return 1
    fi
    local repo
    repo="$(echo "$remote_url" | sed -E 's#^(git@[^:]+:|https?://[^/]+/)##; s#\.git$##')"

    echo -e "${YELLOW}触发 ${plugin_name} 插件发布...${NC}"
    echo "  仓库: ${repo}"

    if [ -n "$version" ]; then
        echo "  版本: ${version}"
        gh workflow run release.yml \
            --repo "${repo}" \
            --field version="${version}"
    else
        echo "  版本: 自动生成日期版本"
        gh workflow run release.yml \
            --repo "${repo}"
    fi

    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✓ ${plugin_name} 发布 workflow 已触发${NC}"
    else
        echo -e "${RED}✗ ${plugin_name} 发布触发失败${NC}"
        return 1
    fi
    echo ""
}

# 检查 gh CLI
if ! command -v gh &> /dev/null; then
    echo -e "${RED}错误: 需要安装 GitHub CLI (gh)${NC}"
    echo "安装: https://cli.github.com/"
    exit 1
fi

# 检查参数
if [ -z "$1" ]; then
    usage
fi

PLUGIN_NAME="$1"
VERSION="$2"

if [ "$PLUGIN_NAME" = "all" ]; then
    echo -e "${YELLOW}=== 发布所有插件 ===${NC}"
    echo ""
    for p in "${PLUGINS[@]}"; do
        release_plugin "$p" "$VERSION"
    done
    echo -e "${GREEN}=== 所有插件发布 workflow 已触发 ===${NC}"
else
    # 验证插件名
    valid=false
    for p in "${PLUGINS[@]}"; do
        if [ "$PLUGIN_NAME" = "$p" ]; then
            valid=true
            break
        fi
    done

    if [ "$valid" = false ]; then
        echo -e "${RED}错误: 未知插件 '${PLUGIN_NAME}'${NC}"
        echo ""
        usage
    fi

    release_plugin "$PLUGIN_NAME" "$VERSION"
fi
