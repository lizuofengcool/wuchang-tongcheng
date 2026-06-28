import { defineConfig } from 'vite'
import uni from '@dcloudio/vite-plugin-uni'

// uni-app + Vue3 + Vite 配置
export default defineConfig({
  base: '/h5/',
  plugins: [uni()],
  server: {
    port: 5174,
    // H5 开发模式：把 /api 代理到后端，避免 CORS
    proxy: {
      '/api': {
        target: process.env.BACKEND_URL || 'http://localhost:8080',
        changeOrigin: true,
      },
    },
  },
})
