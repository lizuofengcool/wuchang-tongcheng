/** @type {import('next').NextConfig} */
const nextConfig = {
  // 通过统一代理访问时使用 /pc 前缀
  basePath: '/pc',
  // 后端 API 代理：避免开发环境 CORS，生产环境用 nginx 反代
  async rewrites() {
    const backend = process.env.BACKEND_URL || 'http://localhost:8080'
    return [
      {
        source: '/api/:path*',
        destination: `${backend}/api/:path*`,
      },
    ]
  },
  // 图片域名白名单（封面图等来自 MinIO/local 的 URL）
  images: {
    remotePatterns: [
      { protocol: 'http', hostname: '**' },
      { protocol: 'https', hostname: '**' },
    ],
  },
}

export default nextConfig
