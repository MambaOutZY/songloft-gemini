<script setup lang="ts">
import { inject, ref, type Ref } from 'vue'
import type { Lang } from '../../../data/downloads'
import { createT } from '../../../data/landing-i18n'

const props = defineProps<{ code: string; title?: string }>()

const lang = inject<Ref<Lang>>('landingLang')
const t = (k: string) => createT(lang?.value ?? 'zh')(k)

const copied = ref(false)
async function copy() {
  try {
    await navigator.clipboard.writeText(props.code)
    copied.value = true
    setTimeout(() => (copied.value = false), 1600)
  } catch {
    /* clipboard 不可用时静默 */
  }
}
</script>

<template>
  <div class="code-block">
    <div class="code-head">
      <span class="code-title">{{ title }}</span>
      <button class="code-copy" type="button" @click="copy">
        {{ copied ? t('install.copied') : t('install.copy') }}
      </button>
    </div>
    <pre class="code-body"><code>{{ code }}</code></pre>
  </div>
</template>

<style scoped>
.code-block {
  border: 1px solid var(--vp-c-divider);
  border-radius: 12px;
  overflow: hidden;
  background: var(--vp-c-bg-alt);
}
.code-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 8px 12px;
  background: var(--vp-c-bg-soft);
  border-bottom: 1px solid var(--vp-c-divider);
}
.code-title {
  font-size: 13px;
  font-weight: 600;
  color: var(--vp-c-text-2);
}
.code-copy {
  font-size: 12px;
  font-weight: 600;
  padding: 4px 12px;
  border-radius: 999px;
  color: var(--vp-c-brand-1);
  border: 1px solid var(--vp-c-brand-1);
  background: transparent;
  transition: background 0.2s, color 0.2s;
}
.code-copy:hover {
  background: var(--vp-c-brand-1);
  color: #fff;
}
.code-body {
  margin: 0;
  padding: 14px 16px;
  overflow-x: auto;
  font-size: 13px;
  line-height: 1.7;
  font-family: var(--vp-font-family-mono);
  color: var(--vp-c-text-1);
}
.code-body code {
  white-space: pre;
}
</style>
