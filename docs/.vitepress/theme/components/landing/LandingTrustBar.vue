<script setup lang="ts">
import { inject, type Ref } from 'vue'
import type { Lang } from '../../../data/downloads'
import { createT } from '../../../data/landing-i18n'

const lang = inject<Ref<Lang>>('landingLang')!
const t = (k: string) => createT(lang.value)(k)

const platforms = ['Linux', 'macOS', 'Windows', 'Android', 'iOS', 'Web', 'Docker', 'Kodi', 'NAS']
</script>

<template>
  <section class="trust" data-reveal>
    <div class="landing-container">
      <p class="trust-title">{{ t('trust.title') }}</p>
      <div class="trust-row">
        <span v-for="p in platforms" :key="p" class="trust-chip">{{ p }}</span>
      </div>
    </div>
  </section>
</template>

<style scoped>
.trust {
  padding: 8px 0 8px;
}
.trust-title {
  text-align: center;
  font-size: 13px;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: var(--vp-c-text-3);
  margin: 0 0 18px;
}
.trust-row {
  display: flex;
  flex-wrap: wrap;
  justify-content: center;
  gap: 10px 12px;
}
.trust-chip {
  font-size: 14px;
  font-weight: 600;
  color: var(--vp-c-text-2);
  padding: 8px 16px;
  border-radius: 999px;
  border: 1px solid var(--vp-c-divider);
  background: var(--vp-c-bg-soft);
}
</style>
