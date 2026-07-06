<script setup lang="ts">
import { inject, type Ref } from 'vue'
import { pick, type Lang } from '../../../data/downloads'
import type { FeatureRow } from '../../../data/features'
import BrowserFrame from './BrowserFrame.vue'
import DeviceFrame from './DeviceFrame.vue'

defineProps<{ feature: FeatureRow }>()

const lang = inject<Ref<Lang>>('landingLang')!
</script>

<template>
  <div class="feature-row" :class="{ reverse: feature.reverse }" data-reveal>
    <div class="feature-copy">
      <h3 class="feature-title">{{ pick(feature.title, lang) }}</h3>
      <p class="feature-desc">{{ pick(feature.desc, lang) }}</p>
      <ul class="feature-bullets">
        <li v-for="(b, i) in feature.bullets" :key="i">
          <span class="tick" aria-hidden="true">✓</span>{{ pick(b, lang) }}
        </li>
      </ul>
    </div>
    <div class="feature-media">
      <DeviceFrame v-if="feature.frame === 'phone'" :src="feature.image" :alt="pick(feature.title, lang)" />
      <BrowserFrame v-else :src="feature.image" :alt="pick(feature.title, lang)" />
    </div>
  </div>
</template>

<style scoped>
.feature-row {
  display: grid;
  grid-template-columns: minmax(0, 1fr) minmax(0, 1.1fr);
  gap: 48px;
  align-items: center;
  padding: 44px 0;
}
.feature-row.reverse .feature-copy { order: 2; }
.feature-row.reverse .feature-media { order: 1; }
.feature-title {
  font-size: clamp(1.5rem, 2.6vw, 2rem);
  font-weight: 750;
  letter-spacing: -0.01em;
  margin: 0;
  border: none;
  padding: 0;
}
.feature-desc {
  margin: 16px 0 0;
  font-size: 1.05rem;
  line-height: 1.75;
  color: var(--vp-c-text-2);
}
.feature-bullets {
  list-style: none;
  padding: 0;
  margin: 22px 0 0;
  display: flex;
  flex-direction: column;
  gap: 10px;
}
.feature-bullets li {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 0.98rem;
  color: var(--vp-c-text-1);
}
.tick {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 20px;
  height: 20px;
  flex: none;
  border-radius: 50%;
  font-size: 12px;
  font-weight: 700;
  color: #fff;
  background: var(--vp-c-brand-1);
}
.feature-media { max-width: 100%; }
@media (max-width: 880px) {
  .feature-row {
    grid-template-columns: 1fr;
    gap: 28px;
    padding: 32px 0;
  }
  .feature-row.reverse .feature-copy,
  .feature-row.reverse .feature-media { order: initial; }
}
</style>
