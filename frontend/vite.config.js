import { defineConfig, loadEnv } from 'vite'
import vue from '@vitejs/plugin-vue'
import { fileURLToPath, URL } from 'node:url'

// Vite 配置
// envDir 默认为项目根目录，会自动加载 .env / .env.[mode]
export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), '')
  const backendURL = env.VITE_BACKEND_URL || 'http://localhost:8080'
  const basePath = env.VITE_BASE_PATH || '/'

  return {
    base: basePath,
    plugins: [vue()],
    resolve: {
      alias: {
        '@': fileURLToPath(new URL('./src', import.meta.url))
      }
    },
    server: {
      port: 5173,
      host: '0.0.0.0',
      proxy: {
        // 开发环境代理后端 API
        '/api': {
          target: backendURL,
          changeOrigin: true
        },
        // 上传文件代理
        '/uploads': {
          target: backendURL,
          changeOrigin: true
        }
      }
    }
  }
})
