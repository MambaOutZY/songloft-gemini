// 落地页固定文案的中英字典。Landing.vue 按当前 lang 取值，缺失回退中文。
import type { Lang } from './downloads'

export type Dict = Record<string, string>

const zh: Dict = {
  // Hero
  'hero.badge': '自托管 · 开源 · 无遥测',
  'hero.title': '你的音乐，你的服务器',
  'hero.subtitle':
    'Songloft 是面向个人的自托管音乐服务器：数据完全自主、插件化扩展、跨平台客户端，仅管理你合法拥有的音乐。',
  'hero.cta.primary': '快速开始',
  'hero.cta.secondary': '选择安装方式',
  'hero.cta.github': 'GitHub',
  'hero.note': '默认账号 admin / admin，仅供本地测试；对外部署请先设置强密码。',

  // Trust bar
  'trust.title': '一处部署，处处可听',

  // Features section
  'features.eyebrow': '核心能力',
  'features.title': '为「拥有」而生的音乐体验',
  'features.subtitle': '轻量高效的 Go 后端，配合精致的跨平台客户端。',

  // Screenshots gallery
  'gallery.eyebrow': '界面一览',
  'gallery.title': '每一屏都清爽好用',
  'gallery.subtitle': '亮色 / 暗色、桌面 / 移动，随手都好看。',

  // Installer
  'install.eyebrow': '开始使用',
  'install.title': '选择你的安装方式',
  'install.subtitle': '二进制、Docker、Bundle、客户端与大屏，总有一款适合你。',
  'install.security': '安全提示：默认 admin / admin 仅限本地测试。任何对外暴露的部署请通过 ADMIN_USERNAME / ADMIN_PASSWORD 设置强密码。',
  'install.edition': '版本',
  'install.platform': '平台',
  'install.download': '下载',
  'install.copy': '复制',
  'install.copied': '已复制',
  'install.open': '打开',
  'install.docsHint': '需要更详细的步骤？查看',
  'install.docsLink': '快速开始文档',

  // Plugins
  'plugins.eyebrow': '插件生态',
  'plugins.title': '用插件把边界推得更远',
  'plugins.subtitle':
    '基于 QuickJS 沙箱运行的 JS 插件，可扩展音源、元数据、设备控制等能力；权限模型、热更新与健康检查全自动。',
  'plugins.cta.list': '浏览插件合集',
  'plugins.cta.dev': '插件开发指南',
  'plugins.cta.registry': '插件源制作',

  // Compliance
  'compliance.title': '合规与版权',
  'compliance.intro': 'Songloft 是帮助你管理自己合法拥有音乐的工具，请在使用前了解：',
  'compliance.i1': '不内置、不分发任何受版权保护的音乐内容',
  'compliance.i2': '请仅管理你合法拥有的音乐（购买 / 自录 / 公有领域 / CC 授权等）',
  'compliance.i3': '第三方插件由社区维护，音源版权责任由使用者自负',
  'compliance.i4': '仅供个人非商业使用，服务端无任何遥测',
  'compliance.more': '详见',
  'compliance.notice': 'NOTICE',
  'compliance.privacy': '隐私说明',

  // Final CTA
  'cta.title': '准备好拥有自己的音乐服务器了吗？',
  'cta.subtitle': '几分钟即可跑起来，数据始终在你手中。',
  'cta.primary': '立即开始',
  'cta.github': '在 GitHub 上 Star',
}

const en: Dict = {
  'hero.badge': 'Self-hosted · Open source · No telemetry',
  'hero.title': 'Your music, your server',
  'hero.subtitle':
    'Songloft is a self-hosted music server for individuals: fully own your data, extend with plugins, play everywhere — for music you legally own.',
  'hero.cta.primary': 'Get started',
  'hero.cta.secondary': 'Choose install',
  'hero.cta.github': 'GitHub',
  'hero.note': 'Default admin / admin is for local testing only — set a strong password before exposing it.',

  'trust.title': 'Deploy once, listen everywhere',

  'features.eyebrow': 'Core features',
  'features.title': 'A music experience built around ownership',
  'features.subtitle': 'A lightweight Go backend paired with polished cross-platform clients.',

  'gallery.eyebrow': 'Interface',
  'gallery.title': 'Clean and usable on every screen',
  'gallery.subtitle': 'Light or dark, desktop or mobile — it just looks good.',

  'install.eyebrow': 'Get started',
  'install.title': 'Choose how to install',
  'install.subtitle': 'Binary, Docker, Bundle, clients and big screens — there is one for you.',
  'install.security': 'Security: default admin / admin is local-only. For any exposed deployment, set a strong ADMIN_USERNAME / ADMIN_PASSWORD.',
  'install.edition': 'Edition',
  'install.platform': 'Platform',
  'install.download': 'Download',
  'install.copy': 'Copy',
  'install.copied': 'Copied',
  'install.open': 'Open',
  'install.docsHint': 'Need step-by-step details? See the',
  'install.docsLink': 'Quick Start docs',

  'plugins.eyebrow': 'Plugin ecosystem',
  'plugins.title': 'Push the limits with plugins',
  'plugins.subtitle':
    'JS plugins run in a QuickJS sandbox to add audio sources, metadata and device control — permissions, hot-reload and health checks all automatic.',
  'plugins.cta.list': 'Browse plugins',
  'plugins.cta.dev': 'Plugin dev guide',
  'plugins.cta.registry': 'Build a registry',

  'compliance.title': 'Compliance & copyright',
  'compliance.intro': 'Songloft helps you manage music you legally own. Before using it, please note:',
  'compliance.i1': 'It ships and distributes no copyrighted music content',
  'compliance.i2': 'Manage only music you legally own (purchased / self-recorded / public domain / CC)',
  'compliance.i3': 'Third-party plugins are community-maintained; source copyright is the user’s responsibility',
  'compliance.i4': 'For personal, non-commercial use; the server has no telemetry',
  'compliance.more': 'See',
  'compliance.notice': 'NOTICE',
  'compliance.privacy': 'Privacy',

  'cta.title': 'Ready to own your music server?',
  'cta.subtitle': 'Up and running in minutes — your data stays with you.',
  'cta.primary': 'Start now',
  'cta.github': 'Star on GitHub',
}

export const MESSAGES: Record<Lang, Dict> = { zh, en }

export function createT(lang: Lang) {
  const dict = MESSAGES[lang] ?? zh
  return (key: string): string => dict[key] ?? zh[key] ?? key
}
