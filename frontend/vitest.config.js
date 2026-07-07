// Vitest 配置（独立于 vite.config.js，仅用于单元测试）
// 复用 @ -> src 别名，与构建配置保持一致；测试为纯函数/状态逻辑，使用 node 环境
import { defineConfig } from 'vitest/config'
import vue from '@vitejs/plugin-vue'
import { fileURLToPath, URL } from 'node:url'

export default defineConfig({
  plugins: [vue()],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url))
    }
  },
  test: {
    environment: 'node',
    setupFiles: ['./vitest.setup.js'],
    include: ['src/**/*.{test,spec}.js']
  }
})
