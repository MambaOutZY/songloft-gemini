# Songloft 颜色系统规范

本文档说明 Songloft 项目中使用的颜色体系及其使用规范。

## 📚 目录

- [Flutter Material 3 颜色体系](#flutter-material-3-颜色体系)
- [主题配置](#主题配置)
- [主题包系统](#主题包系统)
- [颜色使用规范](#颜色使用规范)
- [响应式主题适配](#响应式主题适配)

---

## Flutter Material 3 颜色体系

Songloft Flutter 前端使用 **Material 3** 设计系统，通过 `ColorScheme.fromSeed` 自动生成完整的颜色方案。

### 核心配置

```dart
// songloft-player/lib/core/theme/app_theme.dart
class AppTheme {
  static const Color _defaultSeedColor = Color(0xFF415F91); // M3 Blue baseline

  static ThemeData lightTheme({
    ScreenType screenType = ScreenType.mobile,
    ThemePack? themePack, // 可选主题包覆盖
  }) {
    return _buildTheme(Brightness.light, screenType, themePack);
  }

  static ThemeData darkTheme({
    ScreenType screenType = ScreenType.mobile,
    ThemePack? themePack,
  }) {
    return _buildTheme(Brightness.dark, screenType, themePack);
  }
}
```

当传入 `themePack` 时，seedColor 会被主题包中定义的颜色替换，从而生成完全不同的调色板。

### 优势

1. **自动配色**：从 seed color 自动生成完整的亮色/暗色配色方案
2. **语义化角色**：`primary`、`secondary`、`tertiary`、`error` 等语义化颜色角色
3. **对比度保证**：Material 3 自动确保文本与背景的对比度符合无障碍标准
4. **一致性**：所有组件自动使用统一的配色方案

### ColorScheme 颜色角色

| 角色 | 用途 | 示例 |
|------|------|------|
| `primary` | 主要操作、强调元素 | 播放按钮、导航选中态 |
| `onPrimary` | primary 上的文本/图标 | 按钮文字 |
| `primaryContainer` | 主色容器背景 | 选中卡片背景 |
| `secondary` | 次要操作 | 辅助按钮 |
| `tertiary` | 第三级强调 | 标签、徽章 |
| `error` | 错误状态 | 删除按钮、错误提示 |
| `surface` | 页面/卡片背景 | Scaffold 背景 |
| `onSurface` | surface 上的文本 | 主要文本 |
| `onSurfaceVariant` | 次要文本 | 副标题、说明文字 |
| `outline` | 边框 | 输入框边框、分割线 |
| `outlineVariant` | 弱化边框 | 列表分割线 |

---

## 主题配置

### 主题模式

Songloft 支持三种主题模式：

- **亮色模式**：明亮的界面风格
- **暗色模式**：护眼的暗色界面
- **跟随系统**：自动跟随操作系统设置

主题切换通过 `ThemeSelector` 组件实现，状态由 `themeModeProvider` 管理。

### 字体配置

```dart
ThemeData(
  fontFamilyFallback: const ['NotoSansSC', 'sans-serif'],
  // ...
)
```

- 默认使用系统字体
- 中文回退到 **Noto Sans SC**（随应用打包）

### 组件主题定制

```dart
ThemeData(
  useMaterial3: true,
  appBarTheme: const AppBarTheme(centerTitle: false, elevation: 0),
  cardTheme: CardThemeData(elevation: 0, shape: RoundedRectangleBorder(...)),
  inputDecorationTheme: InputDecorationTheme(border: OutlineInputBorder(...), filled: true),
  navigationBarTheme: const NavigationBarThemeData(height: 64, ...),
  // ...
)
```

---

## 主题包系统

Songloft 支持通过 `.songloft-theme` 主题包自定义应用的配色和视觉样式。主题模式（亮色/暗色/跟随系统）和主题包相互独立——一个主题包同时定义亮色和暗色两套配色。

### SongloftThemeExtension

通过 `ThemeExtension` 机制注入主题包特有的自定义参数：

```dart
class SongloftThemeExtension extends ThemeExtension<SongloftThemeExtension> {
  final List<Color>? playerGradientColors; // 播放器渐变色
  final double cardRadius;                 // 卡片圆角
  final double controlRadius;              // 控件圆角
  final double navigationRadius;           // 导航圆角
}
```

在组件中使用：

```dart
final ext = Theme.of(context).extension<SongloftThemeExtension>();
if (ext?.playerGradientColors != null) {
  // 使用主题包的渐变色
}
```

### 主题包如何影响颜色系统

当用户激活一个主题包时：

1. **seedColor 被替换**：从主题包的 `light.seedColor` / `dark.seedColor` 生成新的完整调色板
2. **surface/background 可覆盖**：如果主题包指定了 `backgroundColor` 或 `surfaceColor`，会 `copyWith` 覆盖自动生成的值
3. **播放器渐变叠加**：`playerGradient` 定义的颜色以 40% 透明度叠加在封面动态取色之上
4. **圆角半径**：`cardRadius`、`controlRadius`、`navigationRadius` 注入到组件主题中

### 相关文档

- 📖 [主题包制作指南](/theme-pack-guide) — 详细的主题制作流程和最佳实践
- 🎨 [在线主题仓库](https://github.com/songloft-org/songloft-themes) — 社区主题包

---

## 颜色使用规范

### ✅ 推荐使用

#### 1. 通过 Theme 获取颜色

```dart
// 获取 ColorScheme
final colorScheme = Theme.of(context).colorScheme;

// 主色
Container(color: colorScheme.primary)
Text('标题', style: TextStyle(color: colorScheme.onSurface))

// 次要文本
Text('描述', style: TextStyle(color: colorScheme.onSurfaceVariant))

// 错误状态
Icon(Icons.error, color: colorScheme.error)
```

#### 2. 使用 TextTheme

```dart
final textTheme = Theme.of(context).textTheme;

Text('大标题', style: textTheme.headlineMedium)
Text('正文', style: textTheme.bodyLarge)
Text('说明', style: textTheme.bodySmall)
```

#### 3. 使用 Material 组件的内置颜色

```dart
// FilledButton 自动使用 primary 色
FilledButton(onPressed: () {}, child: Text('主要操作'))

// OutlinedButton 自动使用 outline 色
OutlinedButton(onPressed: () {}, child: Text('次要操作'))

// TextButton 自动使用 primary 色
TextButton(onPressed: () {}, child: Text('文本操作'))
```

### ❌ 避免使用

```dart
// 不要硬编码颜色值
Container(color: Color(0xFF415F91))  // ❌

// 不要使用 Colors 常量（不跟随主题）
Text('文本', style: TextStyle(color: Colors.grey))  // ❌

// 应该使用 Theme
Container(color: Theme.of(context).colorScheme.primary)  // ✅
Text('文本', style: TextStyle(color: Theme.of(context).colorScheme.onSurfaceVariant))  // ✅
```

---

## 响应式主题适配

主题根据屏幕类型（Mobile / Tablet / Desktop）动态调整组件尺寸：

### SnackBar

| 屏幕类型 | 样式 |
|---------|------|
| Mobile | 默认浮动样式 |
| Desktop | 固定宽度 480px，居中 |

### FilledButton / OutlinedButton / TextButton

| 屏幕类型 | 最小尺寸 |
|---------|---------|
| Desktop | 88 × 44 |
| Mobile / Tablet | Flutter 框架默认（未定制） |

> 实际代码 (`app_theme.dart`) 中 `filledButtonTheme` / `outlinedButtonTheme` / `textButtonTheme` 三者均以 `isDesktop` 为前置条件：仅在 Desktop 时才设置 theme（Desktop → 88×44），Mobile / Tablet 为 `null`，沿用 Flutter 框架默认尺寸。

### 对话框最大宽度

| 屏幕类型 | 最大宽度 |
|---------|---------|
| Mobile | 300px |
| Tablet | 400px |
| Desktop | 480px |

---

## 封面颜色提取

Songloft 使用 `palette_generator` 库从歌曲封面图片中提取主色调，用于播放器界面的动态配色：

```dart
// songloft-player/lib/core/utils/color_extraction.dart
// 从封面图片提取主色调，应用到播放器背景渐变等场景
```

---

## 更新日志

- **2026-07-31**: 新增主题包系统
  - 支持 `.songloft-theme` 主题包自定义配色、圆角、播放器渐变
  - 新增 `SongloftThemeExtension` 自定义主题扩展
  - `AppTheme.lightTheme()` / `darkTheme()` 接受可选 `ThemePack` 参数
  - 在线主题目录：从 [songloft-themes](https://github.com/songloft-org/songloft-themes) 浏览安装
- **2026-04-14**: 更新为 Flutter Material 3 颜色体系
  - 主前端迁移到 Flutter，使用 `ColorScheme.fromSeed` 自动配色
  - seedColor: M3 Blue baseline (`#415F91`)
  - 新增响应式主题适配（Mobile / Tablet / Desktop）
  - 新增封面颜色提取功能
