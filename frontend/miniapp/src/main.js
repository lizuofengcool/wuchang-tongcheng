import { createSSRApp } from 'vue'
import App from './App.vue'

// uni-app 入口
export function createApp() {
  const app = createSSRApp(App)
  return { app }
}
