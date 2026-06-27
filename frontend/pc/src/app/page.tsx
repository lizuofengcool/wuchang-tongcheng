import Link from 'next/link'
import NewsCard from '@/components/NewsCard'
import { listNews, listCategories } from '@/lib/api'

export const revalidate = 60 // ISR：每 60 秒重新生成

export default async function HomePage() {
  // 服务端并行拉取最新头条和分类
  // 后端不可用时降级为空列表（避免 build/prerender 失败）
  let newsPage = { list: [] as any[], total: 0, page: 1, pageSize: 10 }
  let categories: any[] = []
  try {
    [newsPage, categories] = await Promise.all([
      listNews({ page: 1, pageSize: 12 }),
      listCategories(),
    ])
  } catch {
    // 后端未启动或不可达，渲染空状态
  }

  return (
    <div className="container py-6">
      {/* 分类导航 */}
      <section className="bg-white rounded-lg p-4 mb-6 border border-gray-200">
        <h2 className="font-bold mb-3">分类导航</h2>
        <div className="flex flex-wrap gap-2">
          {categories.length === 0 && (
            <span className="text-sm text-gray-500">暂无分类</span>
          )}
          {categories.map((c) => (
            <Link
              key={c.id}
              href={`/category/${c.id}`}
              className="px-3 py-1.5 text-sm bg-gray-100 rounded hover:bg-brand-50 hover:text-brand-600"
            >
              {c.name}
            </Link>
          ))}
        </div>
      </section>

      {/* 最新头条 */}
      <section>
        <div className="flex items-center justify-between mb-4">
          <h2 className="text-xl font-bold">最新同城头条</h2>
          <Link href="/news" className="text-sm text-brand-600 hover:underline">
            查看全部 →
          </Link>
        </div>

        {newsPage.list.length === 0 ? (
          <div className="bg-white rounded-lg p-12 text-center text-gray-500 border border-gray-200">
            暂无头条内容
          </div>
        ) : (
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">
            {newsPage.list.map((news) => (
              <NewsCard key={news.id} news={news} />
            ))}
          </div>
        )}
      </section>
    </div>
  )
}
