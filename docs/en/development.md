# Documentation Site Development Guide

This document is for developers who want to preview or modify the Songloft documentation site locally.

## Local Preview

```bash
cd docs
npm install
npm run docs:dev
```

By default it listens on `http://localhost:3030`. `docs:dev` first runs `sync` (which syncs the repository root's `README.md` / `CHANGELOG.md` into `docs/quick-start.md` / `docs/changelog.md`) and `fetch-issues` (which pulls issues labeled "文档" from GitHub to generate `docs/issues/*.md`), then starts VitePress.

## Directory Categories

The Markdown files under `docs/` fall into three categories. Determine a file's type before modifying it:

| File | Type | How to Modify |
|------|------|---------|
| `index.md` | Source file | Edit directly |
| `faq.md` | Source file | Edit directly |
| `js-plugin-development-guide.md` | Source file | Edit directly |
| `swagger.json` | Manually maintained | Copy from the main repo's `songloft/docs/swagger.json` |
| `quick-start.md` | Generated at build time | Edit the repository root's `README.md` |
| `changelog.md` | Generated at build time | Edit the repository root's `CHANGELOG.md` |
| `issues/*.md` | Generated at runtime | Edit GitHub issues (labeled "文档") |

All generated files are ignored by `docs/.gitignore`. Do not commit them, and do not edit them manually.

## Sync Scripts

- `scripts/sync-docs.mjs`: Repository root Markdown → `docs/`. To add a new sync item, just append a line to the `syncItems` array.
- `scripts/fetch-issues.mjs`: Calls the GitHub API to fetch issues labeled "文档" and writes them to `docs/issues/*.md`. Depends on the environment variable `VITE_GITHUB_ISSUES_TOKEN` (injected from secrets in CI). If this variable is not set locally, the script clears `docs/issues/` and skips directly; the site can still be built, but the issues page will be empty.

Both scripts are chained into `docs:dev` / `docs:build` by `docs/package.json` and must run **before** `vitepress build`, otherwise VitePress will not pick up the newly generated pages.
