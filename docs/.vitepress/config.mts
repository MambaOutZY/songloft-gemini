import { defineConfig } from 'vitepress'
import AutoSidebar from 'vite-plugin-vitepress-auto-sidebar';
import taskLists from 'markdown-it-task-lists'

export default async () => {
  return defineConfig({
    title: "Songloft",
    description: "Songloft - 自托管个人音乐服务器，支持 JS 插件扩展，跨平台 Flutter 客户端",
    srcExclude: ['repowiki/**'],

    head: [
      ['link', { rel: 'icon', type: 'image/svg+xml', href: '/favicon.svg' }],
      ['meta', { property: 'og:type', content: 'website' }],
      ['meta', { property: 'og:title', content: 'Songloft - 自托管个人音乐服务器' }],
      ['meta', { property: 'og:description', content: '简单、自由、插件化的个人音乐服务器，支持 JS 插件扩展' }],
      ['meta', { property: 'og:image', content: 'https://songloft.hanxi.cc/logo.png' }],
    ],

    themeConfig: {
      logo: '/logo.png',

      nav: [
        { text: '快速开始', link: '/quick-start' },
        { text: '客户端', link: '/issues/8' },
        {
          text: '插件',
          items: [
            { text: '插件列表', link: '/issues/4' },
            { text: '插件开发指南', link: '/js-plugin-development-guide' },
            { text: '插件源制作指南', link: '/plugin_registry' },
          ],
        },
        { text: 'FAQ', link: '/faq' },
        { text: '更新日志', link: '/changelog' },
        {
          text: '更多',
          items: [
            { text: 'API 文档', link: 'https://petstore.swagger.io/?url=https://raw.githubusercontent.com/songloft-org/songloft/refs/heads/main/docs/swagger.json' },
            { text: 'Docker Hub', link: 'https://hub.docker.com/r/songloft/songloft' },
            { text: '隐私说明', link: '/PRIVACY' },
            { text: 'NOTICE', link: '/NOTICE' },
          ],
        },
      ],

      socialLinks: [
        { icon: 'github', link: 'https://github.com/songloft-org/songloft' },
      ],

      footer: {
        message: '基于 <a href="https://github.com/songloft-org/songloft/blob/main/LICENSE">Apache 2.0</a> 协议开源',
        copyright: `Copyright © 2025-${new Date().getFullYear()} <a href="https://github.com/hanxi">涵曦</a>`,
      },

      search: {
        provider: 'local',
      },

      editLink: {
        pattern: 'https://github.com/songloft-org/songloft/issues',
        text: '在 GitHub 上提问',
      },
    },

    sitemap: {
      hostname: 'https://songloft.hanxi.cc',
    },

    lastUpdated: true,

    markdown: {
      lineNumbers: false,
      config: (md) => {
        md.use(taskLists)
      },
    },

    vite: {
      plugins: [
        AutoSidebar({
          path: '.',
          collapsed: true,
          titleFromFile: true,
          ignoreList: ['repowiki'],
        }),
      ],
    },
  })
}
