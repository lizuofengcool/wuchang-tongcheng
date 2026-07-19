// 反向代理：将多个服务通过单端口统一暴露
// 端口从项目根目录 .env 读取，禁止硬编码。
const http = require('http')
const httpProxy = require('http-proxy')
const fs = require('fs')
const path = require('path')

function loadEnv(filePath) {
  const env = {}
  if (!fs.existsSync(filePath)) return env
  const content = fs.readFileSync(filePath, 'utf8')
  for (const raw of content.split('\n')) {
    const line = raw.trim()
    if (!line || line.startsWith('#')) continue
    const idx = line.indexOf('=')
    if (idx === -1) continue
    const key = line.slice(0, idx).trim()
    const value = line.slice(idx + 1).trim()
    env[key] = value
  }
  return env
}

const env = loadEnv(path.resolve(__dirname, '../.env'))

const SERVER_PORT = env.WCTC_SERVER_PORT || '8088'
const ADMIN_PORT = env.WCTC_ADMIN_PORT || '5177'
const PC_PORT = env.WCTC_PC_PORT || '3010'
const H5_PORT = env.WCTC_H5_PORT || '5178'
const PROXY_PORT = parseInt(env.WCTC_PROXY_PORT || '8099', 10)

const proxy = httpProxy.createProxyServer({
  ws: false,
  changeOrigin: true,
  xfwd: true
})

// 服务映射
const routes = {
  '/pc':   `http://localhost:${PC_PORT}`,   // PC 门户 (Next.js)
  '/h5':   `http://localhost:${H5_PORT}`,   // H5 小程序端 (Uni-app)
  '/api':  `http://localhost:${SERVER_PORT}`, // 后端 API
  '/':     `http://localhost:${ADMIN_PORT}`, // 管理后台 (Vue 3)
}

// 错误处理
proxy.on('error', (err, req, res) => {
  console.error('Proxy error:', err.message)
  if (!res.headersSent) {
    res.writeHead(502, { 'Content-Type': 'text/plain' })
  }
  res.end('Service unavailable')
})

const server = http.createServer((req, res) => {
  const url = req.url
  let target = null

  // 匹配路由
  if (url.startsWith('/pc')) {
    target = routes['/pc']
    // Next.js 已配置 basePath: '/pc'，无需重写路径
  } else if (url.startsWith('/h5')) {
    target = routes['/h5']
    // H5 Vite 需要尾部斜杠 /h5/ 才能正确加载
    if (url === '/h5') {
      res.writeHead(301, { Location: '/h5/' })
      res.end()
      return
    }
    // H5 已配置 base: '/h5/'，无需重写路径
  } else if (url.startsWith('/api')) {
    target = routes['/api']
  } else {
    target = routes['/']
  }
  proxy.web(req, res, { target })
})

server.on('upgrade', (req, socket, head) => {
  const url = req.url
  let target = null

  if (url.startsWith('/pc')) {
    target = routes['/pc']
    req.url = url.replace(/^\/pc/, '') || '/'
  } else if (url.startsWith('/h5')) {
    target = routes['/h5']
    req.url = url.replace(/^\/h5/, '') || '/'
  } else {
    target = routes['/']
  }

  if (target) {
    proxy.ws(req, socket, head, { target })
  }
})

server.listen(PROXY_PORT, '0.0.0.0', () => {
  console.log(`Proxy running on http://0.0.0.0:${PROXY_PORT}`)
  console.log(`  /      -> 管理后台 (${ADMIN_PORT})`)
  console.log(`  /pc    -> PC 门户 (${PC_PORT})`)
  console.log(`  /h5    -> H5 小程序 (${H5_PORT})`)
  console.log(`  /api   -> 后端 API (${SERVER_PORT})`)
})
