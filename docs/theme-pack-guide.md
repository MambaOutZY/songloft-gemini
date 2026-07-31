# 主题包制作指南

Songloft 支持通过 `.songloft-theme` 主题包自定义应用的配色、圆角和播放器渐变效果。本文档介绍主题包的格式规范、制作流程和最佳实践。

---

## 概述

主题包是一个 JSON 文件（扩展名 `.songloft-theme`），声明了亮色/暗色模式下的配色方案及 UI 视觉参数。主题包系统基于 Material 3 设计规范：

- 通过一个 **种子色（seedColor）** 自动生成完整的调色板
- 可选覆盖 **背景色** 和 **表面色**
- 独立控制 **播放器渐变**、**圆角半径** 等视觉参数
- 主题模式（亮色/暗色/跟随系统）和主题包**相互独立**——一个主题包同时定义亮色和暗色两套配色

## 完整 Schema

```json
{
  "schemaVersion": 1,
  "id": "作者名.主题名",
  "name": "主题显示名",
  "version": "1.0.0",
  "author": "作者名",
  "description": "简短描述",
  "light": {
    "seedColor": "#6750A4",
    "backgroundColor": "#FFF7FF",
    "surfaceColor": "#FFFFFF"
  },
  "dark": {
    "seedColor": "#D0BCFF",
    "backgroundColor": "#141218",
    "surfaceColor": "#211F26"
  },
  "playerGradient": ["#4A148C", "#1A237E"],
  "cardRadius": 16,
  "controlRadius": 20,
  "navigationRadius": 18
}
```

## 字段详解

### 元信息

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `schemaVersion` | int | ✅ | 固定为 `1`（当前版本） |
| `id` | string | ✅ | 全局唯一 ID，推荐格式 `作者名.主题名`，如 `songloft.neon-night` |
| `name` | string | ✅ | 显示名称 |
| `version` | string | ✅ | 语义化版本号，如 `1.0.0` |
| `author` | string | 否 | 作者名 |
| `description` | string | 否 | 简短描述（建议不超过 50 字） |

### 配色（light / dark）

`light` 和 `dark` 字段分别定义亮色和暗色模式的配色方案，至少需要提供一个。

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `seedColor` | string | ✅ | 种子色，`#RRGGBB` 格式。Material 3 会基于此颜色自动生成 `primary`、`secondary`、`tertiary` 等完整调色板 |
| `backgroundColor` | string | 否 | 覆盖默认背景色，`#RRGGBB` 格式 |
| `surfaceColor` | string | 否 | 覆盖默认表面色（卡片、弹窗背景），`#RRGGBB` 格式 |

#### seedColor 的工作原理

Songloft 使用 Flutter 的 `ColorScheme.fromSeed()` 方法，从 seedColor 自动推导出约 30 个语义化颜色角色（primary、secondary、tertiary、error 等）。你只需选择一个你喜欢的主色调，Material 3 算法会确保所有派生颜色的对比度和和谐性。

::: tip 亮色和暗色的 seedColor 可以不同
亮色模式下通常使用更饱和、更深的种子色；暗色模式下使用更亮的种子色，以获得更好的视觉效果。参考 Neon Night 主题：亮色 `#6750A4`，暗色 `#D0BCFF`。
:::

### 播放器渐变

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `playerGradient` | string[] | 否 | 全屏播放器背景渐变色数组，2 个 `#RRGGBB` 颜色值，从上到下渐变。叠加在封面动态取色之上（alpha 0.4）|

### 圆角半径

| 字段 | 类型 | 必填 | 范围 | 说明 |
|------|------|------|------|------|
| `cardRadius` | number | 否 | 0-100 | 卡片组件的圆角半径 |
| `controlRadius` | number | 否 | 0-100 | 输入框、按钮等控件的圆角半径 |
| `navigationRadius` | number | 否 | 0-100 | 导航指示器的圆角半径 |

## 制作流程

### Step 1：确定配色方案

推荐使用以下工具选择种子色：

- [Material Theme Builder](https://m3.material.io/theme-builder) — Google 官方 M3 主题生成器
- [Coolors](https://coolors.co/) — 配色灵感生成器
- [Adobe Color](https://color.adobe.com/) — 专业色彩工具

选好主色后，在 Material Theme Builder 中输入种子色预览完整调色板效果。

### Step 2：创建主题文件

创建一个 `.songloft-theme` 文件，最简模板：

```json
{
  "schemaVersion": 1,
  "id": "yourname.mytheme",
  "name": "My Theme",
  "version": "1.0.0",
  "author": "Your Name",
  "description": "A brief description",
  "light": {
    "seedColor": "#你的亮色种子色"
  },
  "dark": {
    "seedColor": "#你的暗色种子色"
  }
}
```

只需要一个 `seedColor` 就能生效。`backgroundColor`、`surfaceColor`、`playerGradient`、圆角半径等全部可选。

### Step 3：本地测试

1. 打开 Songloft 客户端
2. 进入 `设置 → 外观 → 主题包 → 导入主题包`
3. 选择你的 `.songloft-theme` 文件
4. 点击已导入的主题包激活
5. 切换亮色/暗色模式确认两套配色都符合预期
6. 打开全屏播放器检查渐变效果

### Step 4：提交到社区

1. Fork [songloft-org/songloft-themes](https://github.com/songloft-org/songloft-themes)
2. 在 `themes/` 下创建主题目录：`themes/你的主题名/`
3. 放入 `theme.songloft-theme` 文件
4. 提交 PR，维护者审核后会更新 `index.json` 并发布

## 最佳实践

### 配色建议

- **种子色选择**：选一个能代表主题风格的色相即可，Material 3 会自动处理饱和度和明度变化
- **亮暗分离**：暗色模式的 seedColor 应比亮色模式更亮，否则暗色模式下 primary 色可能不够醒目
- **背景色谨慎覆盖**：只有在 M3 自动生成的背景色不符合你的设计意图时才需要覆盖。大多数情况下只设置 seedColor 效果就很好
- **渐变色搭配**：`playerGradient` 的两个颜色建议选同色系的深色（暗色），因为会以 40% 透明度叠加在封面取色之上

### ID 命名

- 格式：`作者名.主题名`，全部小写，用连字符分隔单词
- 例如：`songloft.neon-night`、`alice.ocean-breeze`
- ID 一旦发布不要修改，它是主题的唯一标识，修改会导致用户的安装状态丢失

### 版本号

遵循 [语义化版本](https://semver.org/lang/zh-CN/)：

- 修复配色问题：`1.0.0` → `1.0.1`
- 新增渐变效果等功能：`1.0.0` → `1.1.0`
- 大改配色方案：`1.0.0` → `2.0.0`

## 示例

### Neon Night（霓虹之夜）

紫蓝色霓虹风格，适合夜间使用：

```json
{
  "schemaVersion": 1,
  "id": "songloft.neon-night",
  "name": "Neon Night",
  "version": "1.0.0",
  "author": "Songloft",
  "description": "A dark theme with purple and blue neon accents",
  "light": {
    "seedColor": "#6750A4",
    "backgroundColor": "#FFF7FF",
    "surfaceColor": "#FFFFFF"
  },
  "dark": {
    "seedColor": "#D0BCFF",
    "backgroundColor": "#141218",
    "surfaceColor": "#211F26"
  },
  "playerGradient": ["#4A148C", "#1A237E"],
  "cardRadius": 16,
  "controlRadius": 20,
  "navigationRadius": 18
}
```

### Sakura（樱花）

粉色樱花暖色调，温柔浪漫：

```json
{
  "schemaVersion": 1,
  "id": "songloft.sakura",
  "name": "Sakura",
  "version": "1.0.0",
  "author": "Songloft",
  "description": "A warm cherry-blossom theme with pink accents",
  "light": {
    "seedColor": "#D81B60",
    "backgroundColor": "#FFF0F5",
    "surfaceColor": "#FFFFFF"
  },
  "dark": {
    "seedColor": "#F48FB1",
    "backgroundColor": "#1A0A10",
    "surfaceColor": "#261418"
  },
  "playerGradient": ["#880E4F", "#4A148C"],
  "cardRadius": 14,
  "controlRadius": 16,
  "navigationRadius": 14
}
```

### 最简主题（仅种子色）

只用一个种子色，让 Material 3 自动完成所有工作：

```json
{
  "schemaVersion": 1,
  "id": "demo.minimal",
  "name": "Minimal Green",
  "version": "1.0.0",
  "light": {
    "seedColor": "#2E7D32"
  },
  "dark": {
    "seedColor": "#81C784"
  }
}
```

## 故障排除

| 问题 | 原因 | 解决方案 |
|------|------|----------|
| 导入失败，提示格式错误 | JSON 语法错误或缺少必填字段 | 用 JSON 校验工具检查格式，确保 `schemaVersion`、`id`、`name` 存在 |
| 颜色不生效 | 颜色格式不正确 | 必须为 `#RRGGBB` 格式（6 位十六进制，带 `#` 前缀） |
| 圆角无变化 | 圆角值超出范围 | 圆角值须在 0-100 之间 |
| 在线安装失败 | 网络问题或 SHA-256 校验不匹配 | 检查网络连接；如果自建主题源，确保 index.json 中的 SHA-256 与文件一致 |
| 暗色模式颜色太暗 | seedColor 选择过深 | 暗色模式的 seedColor 应选择更亮的色值 |
