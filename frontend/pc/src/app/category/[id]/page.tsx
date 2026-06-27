import Link from 'next/link'
import NewsCard from '@/components/NewsCard'
import { listNews, listCategories } from '@/lib/api'
import { notFound } from 'next/navigation'

export const revalidate = 60

interface PageProps {
  params: { id: string }
  searchParams: { page?: string }
}

export default async function CategoryPage({ params, searchParams }: PageProps) {
  const categoryId = Number(params.id)
  if (Number.isNaN(categoryId)) notFound()

  const page = Number(searchParams.page) || 1

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

  const current = categories.find((c) => c.id === categoryId)

  return (
    <div className="container py-6">
      <nav className="text-sm text-gray-500 mb-4">
        <Link href="/" className="hover:text-brand-600">首页</Link>
        <span className="mx-2">/</span>
        <span>分类：{current?.name || '未知'}</span>
      </nav>

      <h1 className="text-2xl font-bold mb-4">
        {current?.name || '分类'} - 头条列表
      </h1>

      {/* 其他分类切换 */}
      <div className="bg-white rounded-lg p-3 mb-4 border border-gray-200 flex flex-wrap gap-2">
        <Link
          href="/news"
          className="px-3 py-1 text-sm bg-gray-100 rounded hover:bg-brand-50"
        >
          ← 全部头条
        </Link>
        {categories.map((c) => (
          <Link
            key={c.id}
            href={`/category/${c.id}`}
            className={`px-3 py-1 text-sm rounded ${
              c.id === categoryId ? 'bg-brand-600 text-white' : 'bg-gray-100 hover:bg-brand-50'
            }`}
          >
            {c.name}
          </Link>
        ))}
      </div>

      {newsPage.list.length === 0 ? (
        <div className="bg-white rounded-lg p-12 text-center text-gray-500 border border-gray-200">
          该分类下暂无头条
        </div>
      ) : (
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">
          {newsPage.list.map((news) => (
            <NewsCard key={news.id} news={news} />
          ))}
        </div>
      )}
    </div>
  )
}
