import Link from 'next/link'
import NewsCard from '@/components/NewsCard'
import { searchNews } from '@/lib/api'

export const revalidate = 30

interface SearchParams {
  keyword?: string
  page?: string
  category_id?: string
}

export default async function SearchPage({
  searchParams,
}: {
  searchParams: SearchParams
}) {
  const keyword = searchParams.keyword || ''
  const page = Number(searchParams.page) || 1
  const categoryId = searchParams.category_id ? Number(searchParams.category_id) : undefined

  let result = null
  let error = ''
  if (keyword) {
    try {
      result = await searchNews({ keyword, page, pageSize: 12, categoryId })
    } catch (e: any) {
      error = e.message || '搜索失败'
    }
  }

  return (
    <div className="container py-6">
      <h1 className="text-2xl font-bold mb-4">
        {keyword ? `搜索：${keyword}` : '全文搜索'}
      </h1>

      {!keyword && (
        <div className="bg-white rounded-lg p-12 text-center text-gray-500 border border-gray-200">
          请输入搜索关键词
        </div>
      )}

      {error && (
        <div className="bg-red-50 rounded-lg p-4 text-red-600 border border-red-200">
          {error}
        </div>
      )}

      {result && (
        <>
          {result.list.length === 0 ? (
            <div className="bg-white rounded-lg p-12 text-center text-gray-500 border border-gray-200">
              未找到与「{keyword}」相关的头条
            </div>
          ) : (
            <>
              <p className="text-sm text-gray-500 mb-4">
                共找到 {result.total} 条结果
              </p>
              <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">
                {result.list.map((news) => (
                  <NewsCard key={news.id} news={news} />
                ))}
              </div>

              {/* 分页 */}
              {result.total > result.pageSize && (
                <div className="flex items-center justify-center gap-4 mt-8">
                  {page > 1 && (
                    <Link
                      href={`/news/search?keyword=${encodeURIComponent(keyword)}&page=${page - 1}`}
                      className="px-3 py-1.5 text-sm border rounded hover:bg-gray-50"
                    >
                      上一页
                    </Link>
                  )}
                  <span className="text-sm text-gray-600">
                    第 {page} 页（共 {Math.ceil(result.total / result.pageSize)} 页）
                  </span>
                  {page * result.pageSize < result.total && (
                    <Link
                      href={`/news/search?keyword=${encodeURIComponent(keyword)}&page=${page + 1}`}
                      className="px-3 py-1.5 text-sm border rounded hover:bg-gray-50"
                    >
                      下一页
                    </Link>
                  )}
                </div>
              )}
            </>
          )}
        </>
      )}
    </div>
  )
}
