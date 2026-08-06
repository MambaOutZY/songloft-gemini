import DefaultTheme from 'vitepress/theme'
import { h } from 'vue'
import Landing from './components/Landing.vue'
import CopyForLLM from './components/CopyForLLM.vue'
import './custom.css'

export default {
  extends: DefaultTheme,
  Layout: () => {
    return h(DefaultTheme.Layout, null, {
      'doc-before': () => h(CopyForLLM),
    })
  },
  enhanceApp({ app }) {
    // 落地页顶层组件（index.md / en/index.md 中以 <Landing /> 使用）
    app.component('Landing', Landing)
  },
}
