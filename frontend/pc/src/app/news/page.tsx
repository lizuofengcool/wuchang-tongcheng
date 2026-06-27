import Link from 'next/link'
import NewsCard from '@/components/NewsCard'
import { listNews, listCategories } from '@/lib/api'

export const revalidate = 60

interface SearchParams {
  page?: string
  category_id?: string
}

export default async function NewsListPage({
  searchParams,
}: {
  searchParams: SearchParams
}) {
  const page = Number(searchParams.page) || 1
  const categoryId = searchParams.category_id ? Number(searchParams.category_id) : undefined

  let newsPage = { list: [] as any[], total: 0, page, pageSize: 12 }
  let categories: any[] = []
  try {
    [newsPage, categories] = await Promise.all([
      listNews({ page, pageSize: 12, categoryId }),
      listCategories(),
    ])
  } catch {
    // 后端不可达，渲染空状态
  }

  const totalPages = Math.ceil(newsPage.total / newsPage.pageSize)

  return (
    <div className="container py-6">
      <h1 className="text-2xl font-bold mb-4">同城头条</h1>

      {/* 分类筛选 */}
      <div className="bg-white rounded-lg p-3 mb-4 border border-gray-200 flex flex-wrap gap-2 items-center">
        <Link
          href="/news"
          className={`px-3 py-1 text-sm rounded ${
            !categoryId ? 'bg-brand-600 text-white' : 'bg-gray-100 hover:bg-brand-50'
          }`}
        >
          全部
        </Link>
        {categories.map((c) => (
          <Link
            key={c.id}
            href={`/news?category_id=${c.id}`}
            className={`px-3 py-1 text-sm rounded ${
              categoryId === c.id ? 'bg-brand-600 text-white' : 'bg-gray-100 hover:bg-brand-50'
            }`}
          >
            {c.name}
          </Link>
        ))}
      </div>

      {/* 列表 */}
      {newsPage.list.length === 0 ? (
        <div className="bg-white rounded-lg p-12 text-center text-gray-500 border border-gray-200">
          暂无头条
        </div>
      ) : (
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">
          {newsPage.list.map((news) => (
            <NewsCard key={news.id} news={news} />
          ))}
        </div>
      )}

      {/* 分页 */}
      {totalPages > 1 && (
        <div className="flex items-center justify-center gap-2 mt-8">
          {page > 1 && (
            <Link
              href={`/news?page=${page - 1}${categoryId ? `&category_id=${categoryId}` : ''}`}
              className="px-3 py-1.5 text-sm border rounded hover:bg-gray-50"
            >
              上一页
            </Link>
          )}
          <span className="text-sm text-gray-600">
            第 {page} / {totalPages} 页（共 {newsPage.total} 条）
          </span>
          {page < totalPages && (
            <Link
              href={`/news?page=${page + 1}${categoryId ? `&category_id=${categoryId}` : ''}`}
              className="px-3 py-1.5 text-sm border rounded hover:bg-gray-50"
            >
              下一页
            </Link>
          )}
        </div>
      )}
    </div>
  )
}
