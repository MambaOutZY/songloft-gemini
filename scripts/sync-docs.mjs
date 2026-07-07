#!/usr/bin/env node
// 将仓库根下的 Markdown 文件同步到 docs/ 指定位置，供 VitePress 构建使用。
// 所有拷贝目标均被 docs/.gitignore 忽略，请勿手动编辑。
//
// 同步时会把根目录视角的相对链接重写为 docs/ 视角的链接，
// 例如 README.md 中的 (./docs/foo.md) → (./foo.md)、(CHANGELOG.md) → (./changelog.md)、
// (LICENSE) → 指向 GitHub 的绝对 URL。

import { existsSync, mkdirSync, readFileSync, writeFileSync } from 'node:fs';
import { dirname, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);
const repoRoot = resolve(__dirname, '..');

const REPO_BLOB_BASE = 'https://github.com/songloft-org/songloft/blob/main';

const syncItems = [
  // 中文（根 → docs/ 根，root locale）
  { from: 'README.md',    to: 'docs/quick-start.md' },
  { from: 'CHANGELOG.md', to: 'docs/changelog.md' },
  { from: 'NOTICE',       to: 'docs/NOTICE.md' },
  { from: 'PRIVACY.md',   to: 'docs/PRIVACY.md' },
  // 英文（*.en 源 → docs/en/，en locale）。CHANGELOG 不翻译，故英文侧无 changelog 页，
  // en 模式下 rewriteLinks 会把 CHANGELOG.md 链接指向 GitHub 绝对 URL，避免死链。
  { from: 'README.en.md',  to: 'docs/en/quick-start.md', en: true },
  { from: 'NOTICE.en',     to: 'docs/en/NOTICE.md',      en: true },
  { from: 'PRIVACY.en.md', to: 'docs/en/PRIVACY.md',     en: true },
];

// markdown 链接重写：把 (...) 中的链接（不含 http/https/锚点）按规则改写。
// en=true 时用于 docs/en/ 目标：CHANGELOG 无英文页，改为指向 GitHub 绝对链接。
function rewriteLinks(content, { en = false } = {}) {
  return content.replace(/(\]\()([^)\s]+)(\))/g, (match, open, link, close) => {
    if (/^(https?:|mailto:|#)/i.test(link)) return match;

    let path = link;
    let suffix = '';
    const hash = path.indexOf('#');
    if (hash >= 0) {
      suffix = path.slice(hash);
      path = path.slice(0, hash);
    }

    // en 目标位于 docs/en/，源里的 docs/en/xxx.md 需先剥成 ./xxx.md，
    // 否则会被下面的通用规则改写成 ./en/xxx.md，从 docs/en/quick-start.md 视角多出一层 en/ 导致死链。
    if (en) {
      path = path.replace(/^\.\/docs\/en\//, './');
      path = path.replace(/^docs\/en\//, './');
    }
    // ./docs/xxx.md → ./xxx.md
    path = path.replace(/^\.\/docs\//, './');
    // docs/xxx.md → ./xxx.md
    path = path.replace(/^docs\//, './');

    // 仓库根的特殊文件 → GitHub 绝对 URL
    if (/^(LICENSE|Dockerfile|Makefile|go\.mod|go\.sum|main\.go)(\/|$)/i.test(path)) {
      return `${open}${REPO_BLOB_BASE}/${path}${suffix}${close}`;
    }

    // CHANGELOG.md：中文侧对齐到 ./changelog.md；英文侧无该页，改指 GitHub。
    if (path === 'CHANGELOG.md' || path === './CHANGELOG.md') {
      return en
        ? `${open}${REPO_BLOB_BASE}/CHANGELOG.md${suffix}${close}`
        : `${open}./changelog.md${suffix}${close}`;
    }
    // README.md → ./quick-start.md（两侧同名，en 侧解析到 docs/en/quick-start.md）
    if (path === 'README.md' || path === './README.md') {
      return `${open}./quick-start.md${suffix}${close}`;
    }

    return `${open}${path}${suffix}${close}`;
  });
}

let failed = false;

for (const { from, to, en = false } of syncItems) {
  const src = resolve(repoRoot, from);
  const dst = resolve(repoRoot, to);

  if (!existsSync(src)) {
    console.error(`[sync-docs] source file not found: ${src}`);
    failed = true;
    continue;
  }

  try {
    const content = readFileSync(src, 'utf8');
    const rewritten = rewriteLinks(content, { en });
    mkdirSync(dirname(dst), { recursive: true });
    writeFileSync(dst, rewritten, 'utf8');
    console.log(`[sync-docs] ${from} -> ${to}`);
  } catch (err) {
    console.error(`[sync-docs] failed to copy ${from} -> ${to}:`, err);
    failed = true;
  }
}

if (failed) {
  process.exit(1);
}
