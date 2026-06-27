export default function Footer() {
  return (
    <footer className="bg-gray-800 text-gray-300 mt-12">
      <div className="container py-8 text-sm">
        <div className="grid grid-cols-1 md:grid-cols-3 gap-8">
          <div>
            <h4 className="text-white font-bold mb-2">五常同城</h4>
            <p className="text-xs leading-relaxed">
              面向五常市的本地生活服务平台，提供分类信息、同城头条、商家服务。
            </p>
          </div>
          <div>
            <h4 className="text-white font-bold mb-2">快速入口</h4>
            <ul className="space-y-1 text-xs">
              <li><a href="/" className="hover:text-white">首页</a></li>
              <li><a href="/news" className="hover:text-white">同城头条</a></li>
              <li><a href="/news/search" className="hover:text-white">搜索</a></li>
            </ul>
          </div>
          <div>
            <h4 className="text-white font-bold mb-2">联系我们</h4>
            <p className="text-xs">合作邮箱：contact@wuchang.com</p>
            <p className="text-xs mt-1">© {new Date().getFullYear()} 五常同城</p>
          </div>
        </div>
      </div>
    </footer>
  )
}
