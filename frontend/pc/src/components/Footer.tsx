import Link from 'next/link'

// 分站占位：后续接入分站中台后改为动态拉取
const STATIONS = [
  { id: 'wuchang', name: '五常' },
  { id: 'harbin', name: '哈尔滨' },
  { id: 'shangzhi', name: '尚志' },
]

export default function Footer() {
  const year = new Date().getFullYear()
  return (
    <footer className="bg-gray-800 text-gray-300 mt-12">
      <div className="container py-8 text-sm">
        <div className="grid grid-cols-1 md:grid-cols-4 gap-8">
          <div>
            <h4 className="text-white font-bold mb-2">五常同城</h4>
            <p className="text-xs leading-relaxed">
              面向五常市的本地生活服务平台，提供二手、招聘、房产、黄页、商城、头条等分类信息。
            </p>
          </div>
          <div>
            <h4 className="text-white font-bold mb-2">快速入口</h4>
            <ul className="space-y-1 text-xs">
              <li><Link href="/" className="hover:text-white">首页</Link></li>
              <li><Link href="/ershou" className="hover:text-white">二手交易</Link></li>
              <li><Link href="/job" className="hover:text-white">招聘求职</Link></li>
              <li><Link href="/fang" className="hover:text-white">房屋租售</Link></li>
              <li><Link href="/news" className="hover:text-white">同城头条</Link></li>
              <li><Link href="/news/search" className="hover:text-white">搜索</Link></li>
            </ul>
          </div>
          <div>
            <h4 className="text-white font-bold mb-2">分站切换</h4>
            <ul className="space-y-1 text-xs">
              {STATIONS.map((s) => (
                <li key={s.id}>
                  <a href={`/?station=${s.id}`} className="hover:text-white">{s.name}站</a>
                </li>
              ))}
            </ul>
            <p className="text-xs text-gray-500 mt-2">更多分站陆续开通</p>
          </div>
          <div>
            <h4 className="text-white font-bold mb-2">APP 下载</h4>
            <div className="flex items-center gap-4">
              {/* 二维码占位：实际部署时替换为 APP 下载二维码 */}
              <div className="w-20 h-20 bg-white rounded flex items-center justify-center text-gray-400 text-xs">
                APP 下载
              </div>
              <div className="text-xs text-gray-400">
                <p>扫码下载</p>
                <p className="mt-1">五常同城 APP</p>
                <p className="mt-1">iOS / Android</p>
              </div>
            </div>
            <h4 className="text-white font-bold mt-4 mb-2">联系我们</h4>
            <p className="text-xs">合作邮箱：contact@wuchang.com</p>
          </div>
        </div>

        <div className="border-t border-gray-700 mt-6 pt-4 text-xs text-gray-400 flex flex-col md:flex-row justify-between items-center gap-2">
          <p>© {year} 五常同城 本地生活服务平台</p>
          <div className="flex items-center gap-4">
            <a
              href="https://beian.miit.gov.cn/"
              target="_blank"
              rel="noopener noreferrer"
              className="hover:text-white"
            >
              黑ICP备XXXXXXXX号-1
            </a>
            <a
              href="http://www.beian.gov.cn/"
              target="_blank"
              rel="noopener noreferrer"
              className="hover:text-white"
            >
              黑公网安备XXXXXXXXXXXX号
            </a>
          </div>
        </div>
      </div>
    </footer>
  )
}
