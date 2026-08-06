<script setup lang="ts">
import { useData } from 'vitepress'
import { onMounted, onUnmounted, ref } from 'vue'
import {
  defaultAiProviders,
  useCopyOrDownloadAsMarkdownButtons,
} from 'vitepress-plugin-llms/vitepress-components'

const { lang } = useData()

const strings = {
  'zh-Hans': {
    copyPage: '复制本页',
    copied: '已复制',
    viewMarkdown: '查看 Markdown',
    openIn: (name: string) => `在 ${name} 中打开`,
  },
  en: {
    copyPage: 'Copy page',
    copied: 'Copied',
    viewMarkdown: 'View as Markdown',
    openIn: (name: string) => `Open in ${name}`,
  },
}

const t = strings[lang.value as keyof typeof strings] ?? strings.en

const { aiProviders, copied, copyAsMarkdown, openInAI, viewAsMarkdown } =
  useCopyOrDownloadAsMarkdownButtons({
    aiProviders: [...defaultAiProviders, { name: 'Gemini', url: 'https://gemini.google.com/app?q=' }],
  })

const open = ref(false)
const root = ref<HTMLElement>()

function toggle(): void {
  open.value = !open.value
}

function closeMenu(): void {
  open.value = false
}

async function onCopy(): Promise<void> {
  await copyAsMarkdown()
  closeMenu()
}

function onView(): void {
  viewAsMarkdown()
  closeMenu()
}

function onOpenAI(provider: (typeof aiProviders)[number]): void {
  openInAI(provider)
  closeMenu()
}

function onClickOutside(event: MouseEvent): void {
  if (root.value && !root.value.contains(event.target as Node)) closeMenu()
}

onMounted(() => document.addEventListener('click', onClickOutside))
onUnmounted(() => document.removeEventListener('click', onClickOutside))
</script>

<template>
  <div class="copy-for-llm" ref="root">
    <div class="trigger">
      <button class="copy-btn" @click="onCopy">
        <svg v-if="!copied" class="icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <rect x="8" y="8" width="12" height="12" rx="2" />
          <path d="M4 16V5a1 1 0 0 1 1-1h11" />
        </svg>
        <svg v-else class="icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M4 12.5l4.5 4.5L20 6" stroke-linecap="round" stroke-linejoin="round" />
        </svg>
        <span>{{ copied ? t.copied : t.copyPage }}</span>
      </button>
      <span class="divider" />
      <button class="chevron-btn" @click.stop="toggle">
        <svg class="icon chevron" :class="{ open }" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M6 9l6 6 6-6" stroke-linecap="round" stroke-linejoin="round" />
        </svg>
      </button>
    </div>

    <div v-if="open" class="menu">
      <button class="menu-item" @click="onView">
        <span>{{ t.viewMarkdown }}</span>
        <svg class="icon ext" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M7 17L17 7M7 7h10v10" stroke-linecap="round" stroke-linejoin="round" />
        </svg>
      </button>
      <button v-for="provider in aiProviders" :key="provider.name" class="menu-item" @click="onOpenAI(provider)">
        <span>{{ t.openIn(provider.name) }}</span>
        <svg class="icon ext" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M7 17L17 7M7 7h10v10" stroke-linecap="round" stroke-linejoin="round" />
        </svg>
      </button>
    </div>
  </div>
</template>

<style scoped>
.copy-for-llm {
  position: relative;
  display: flex;
  justify-content: flex-end;
  margin-bottom: 16px;
}

.trigger {
  display: flex;
  align-items: stretch;
  border: 1px solid var(--vp-c-divider);
  border-radius: 6px;
  overflow: hidden;
}

.copy-btn,
.chevron-btn,
.menu-item {
  background: transparent;
  border: none;
  color: var(--vp-c-text-1);
  cursor: pointer;
}

.copy-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 6px 12px;
  font-size: 13px;
  white-space: nowrap;
}

.divider {
  width: 1px;
  align-self: stretch;
  margin: 6px 0;
  background: var(--vp-c-divider);
}

.chevron-btn {
  display: flex;
  align-items: center;
  padding: 0 8px;
}

.copy-btn:hover,
.chevron-btn:hover {
  background: var(--vp-c-bg-soft);
  color: var(--vp-c-brand-1);
}

.icon {
  width: 15px;
  height: 15px;
  flex-shrink: 0;
}

.chevron {
  transition: transform 0.2s ease;
}

.chevron.open {
  transform: rotate(180deg);
}

.menu {
  position: absolute;
  top: calc(100% + 4px);
  right: 0;
  min-width: 200px;
  background: var(--vp-c-bg-elv);
  border: 1px solid var(--vp-c-divider);
  border-radius: 8px;
  box-shadow: 0 8px 16px rgba(0, 0, 0, 0.15);
  overflow: hidden;
  z-index: 20;
}

.menu-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  width: 100%;
  padding: 8px 12px;
  font-size: 13px;
  text-align: left;
}

.menu-item:hover {
  background: var(--vp-c-bg-soft);
  color: var(--vp-c-brand-1);
}

.ext {
  opacity: 0.6;
}
</style>
