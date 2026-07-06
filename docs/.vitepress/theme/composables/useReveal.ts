import { onMounted, onUnmounted } from 'vue'

// 滚动淡入：为带 [data-reveal] 的元素在进入视口时加 .revealed。
// 无 JS / SSR 时内容默认可见；仅在挂载后（.reveal-ready）才启用初始隐藏态；
// 尊重 prefers-reduced-motion。
export function useReveal() {
  let io: IntersectionObserver | null = null

  onMounted(() => {
    const root = document.documentElement
    root.classList.add('reveal-ready')

    const els = Array.from(document.querySelectorAll<HTMLElement>('[data-reveal]'))
    const reduced = window.matchMedia('(prefers-reduced-motion: reduce)').matches
    if (reduced || !('IntersectionObserver' in window)) {
      els.forEach((e) => e.classList.add('revealed'))
      return
    }

    io = new IntersectionObserver(
      (entries) => {
        for (const entry of entries) {
          if (entry.isIntersecting) {
            entry.target.classList.add('revealed')
            io?.unobserve(entry.target)
          }
        }
      },
      { threshold: 0.12, rootMargin: '0px 0px -8% 0px' }
    )
    els.forEach((e) => io!.observe(e))
  })

  onUnmounted(() => io?.disconnect())
}
