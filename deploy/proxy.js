// 反向代理：将多个服务通过单端口 5173 暴露
const http = require('http')
const httpProxy = require('http-proxy')

const proxy = httpProxy.createProxyServer({
  ws: false,
  changeOrigin: true,
  xfwd: true
})

// 服务映射
const routes = {
  '/pc':   'http://localhost:3000',   // PC 门户 (Next.js)
  '/h5':   'http://localhost:5174',   // H5 小程序端 (Uni-app)
  '/api':  'http://localhost:8080',   // 后端 API
  '/':     'http://localhost:5175',   // 管理后台 (Vue 3)
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
  let pathRewrite = url

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

const PORT = 5173
server.listen(PORT, '0.0.0.0', () => {
  console.log(`Proxy running on http://0.0.0.0:${PORT}`)
  console.log('  /      → 管理后台 (5175)')
  console.log('  /pc    → PC 门户 (3000)')
  console.log('  /h5    → H5 小程序 (5174)')
  console.log('  /api   → 后端 API (8080)')
})