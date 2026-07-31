# Theme Pack Development Guide

Songloft supports customizing the app's color scheme, border radii, and player gradient effects through `.songloft-theme` theme packs. This guide covers the format specification, creation workflow, and best practices.

---

## Overview

A theme pack is a JSON file (with `.songloft-theme` extension) that declares color schemes for light/dark modes and UI visual parameters. The theme pack system is built on Material 3 design specifications:

- A single **seed color** automatically generates the complete color palette
- Optional overrides for **background** and **surface** colors
- Independent control over **player gradient**, **border radii**, and other visual parameters
- Theme mode (light/dark/system) and theme packs are **independent** — one theme pack defines both light and dark color schemes

## Full Schema

```json
{
  "schemaVersion": 1,
  "id": "author.theme-name",
  "name": "Theme Name",
  "version": "1.0.0",
  "author": "Author Name",
  "description": "Short description",
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

## Field Reference

### Metadata

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `schemaVersion` | int | ✅ | Must be `1` (current version) |
| `id` | string | ✅ | Globally unique ID, format: `author.theme-name`, e.g. `songloft.neon-night` |
| `name` | string | ✅ | Display name |
| `version` | string | ✅ | Semantic version, e.g. `1.0.0` |
| `author` | string | No | Author name |
| `description` | string | No | Short description (max ~50 characters recommended) |

### Colors (light / dark)

`light` and `dark` define color schemes for each mode. At least one must be provided.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `seedColor` | string | ✅ | Seed color in `#RRGGBB` format. Material 3 generates the full `primary`, `secondary`, `tertiary` palette from this color |
| `backgroundColor` | string | No | Override default background color, `#RRGGBB` format |
| `surfaceColor` | string | No | Override default surface color (cards, dialogs), `#RRGGBB` format |

#### How seedColor Works

Songloft uses Flutter's `ColorScheme.fromSeed()` method, which derives ~30 semantic color roles (primary, secondary, tertiary, error, etc.) from the seed color. You only need to pick a main hue — the Material 3 algorithm ensures contrast ratios and color harmony across all derived colors.

::: tip Light and dark can use different seed colors
Light mode typically uses a more saturated, deeper seed color; dark mode uses a brighter one for better visual results. See the Neon Night theme: light `#6750A4`, dark `#D0BCFF`.
:::

### Player Gradient

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `playerGradient` | string[] | No | Full-screen player background gradient, array of 2 `#RRGGBB` colors, top-to-bottom. Layered above cover art color extraction at 40% opacity |

### Border Radii

| Field | Type | Required | Range | Description |
|-------|------|----------|-------|-------------|
| `cardRadius` | number | No | 0-100 | Border radius for cards |
| `controlRadius` | number | No | 0-100 | Border radius for inputs, buttons, controls |
| `navigationRadius` | number | No | 0-100 | Border radius for navigation indicators |

## Creation Workflow

### Step 1: Choose a Color Scheme

Recommended tools for picking seed colors:

- [Material Theme Builder](https://m3.material.io/theme-builder) — Official M3 theme generator by Google
- [Coolors](https://coolors.co/) — Color palette inspiration
- [Adobe Color](https://color.adobe.com/) — Professional color tool

Enter your seed color in Material Theme Builder to preview the full palette.

### Step 2: Create the Theme File

Create a `.songloft-theme` file. Minimal template:

```json
{
  "schemaVersion": 1,
  "id": "yourname.mytheme",
  "name": "My Theme",
  "version": "1.0.0",
  "author": "Your Name",
  "description": "A brief description",
  "light": {
    "seedColor": "#your-light-seed-color"
  },
  "dark": {
    "seedColor": "#your-dark-seed-color"
  }
}
```

A single `seedColor` is all you need. `backgroundColor`, `surfaceColor`, `playerGradient`, and radii are all optional.

### Step 3: Test Locally

1. Open the Songloft app
2. Go to `Settings → Appearance → Theme Packs → Import`
3. Select your `.songloft-theme` file
4. Tap the imported theme to activate it
5. Toggle light/dark mode to verify both color schemes
6. Open the full-screen player to check gradient effects

### Step 4: Submit to the Community

1. Fork [songloft-org/songloft-themes](https://github.com/songloft-org/songloft-themes)
2. Create a directory under `themes/`: `themes/your-theme-name/`
3. Add your `theme.songloft-theme` file
4. Submit a PR — maintainers will update `index.json` after review

## Best Practices

### Color Tips

- **Seed color**: Pick a hue that represents your theme's mood — Material 3 handles saturation and lightness automatically
- **Light vs dark**: Dark mode seedColor should be brighter than light mode, otherwise the primary color won't be prominent enough
- **Override cautiously**: Only override `backgroundColor`/`surfaceColor` when M3's auto-generated colors don't match your design intent. In most cases, seedColor alone works great
- **Gradient colors**: Use dark, same-hue colors for `playerGradient` — they're layered at 40% opacity over the cover art's extracted colors

### ID Naming

- Format: `author.theme-name`, all lowercase, hyphen-separated words
- Examples: `songloft.neon-night`, `alice.ocean-breeze`
- Never change the ID after publishing — it's the unique identifier. Changing it causes users to lose their install state

### Versioning

Follow [Semantic Versioning](https://semver.org/):

- Fix color issues: `1.0.0` → `1.0.1`
- Add gradient effects: `1.0.0` → `1.1.0`
- Major color scheme redesign: `1.0.0` → `2.0.0`

## Examples

### Neon Night

Purple and blue neon style, great for nighttime:

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

### Minimal Theme (seed color only)

Let Material 3 do all the work with just a seed color:

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

## Troubleshooting

| Problem | Cause | Solution |
|---------|-------|----------|
| Import fails with format error | JSON syntax error or missing required fields | Validate JSON syntax, ensure `schemaVersion`, `id`, `name` are present |
| Colors don't apply | Incorrect color format | Must be `#RRGGBB` format (6-digit hex with `#` prefix) |
| Radii unchanged | Radius value out of range | Must be between 0-100 |
| Online install fails | Network issue or SHA-256 mismatch | Check network; for custom sources, ensure SHA-256 in index.json matches the file |
| Dark mode too dark | seedColor too deep | Use a brighter seed color for dark mode |
