import DefaultTheme from 'vitepress/theme'
import Landing from './components/Landing.vue'
import './custom.css'

export default {
  extends: DefaultTheme,
  enhanceApp({ app }) {
    // 落地页顶层组件（index.md / en/index.md 中以 <Landing /> 使用）
    app.component('Landing', Landing)
  },
}
