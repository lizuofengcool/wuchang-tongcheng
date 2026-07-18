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
    },
    build: {
      // 分包：将第三方依赖拆分为独立 chunk，避免单个 vendor chunk 过大
      rollupOptions: {
        output: {
          manualChunks: {
            // Vue 核心生态：vue / vue-router / pinia
            'vue-vendor': ['vue', 'vue-router', 'pinia'],
            // Element Plus 组件库
            'element-plus': ['element-plus'],
            // Element Plus 图标包
            'element-icons': ['@element-plus/icons-vue'],
            // HTTP 客户端
            'axios': ['axios']
          }
        }
      },
      // 分包后 element-plus 全量注册约 900KB（gzip 293KB）属合理范围，
      // 仅当单 chunk 超过 1MB 才告警
      chunkSizeWarningLimit: 1000
    }
  }
})
