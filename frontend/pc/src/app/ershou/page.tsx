import Link from 'next/link'
import type { Metadata } from 'next'
import { listErshous } from '@/lib/api'

export const revalidate = 60

export const metadata: Metadata = {
  title: '二手交易',
  description:
    '五常同城二手交易频道，提供本地二手物品买卖信息，包括数码、家电、家具、衣物、图书等。',
  keywords: ['五常二手', '二手交易', '二手市场', '五常同城二手'],
  openGraph: {
    title: '五常同城二手交易',
    description: '本地二手物品买卖信息',
    type: 'website',
  },
}

interface SearchParams {
  page?: string
  keyword?: string
}

export default async function ErshouListPage({
  searchParams,
}: {
  searchParams: SearchParams
}) {
  const page = Number(searchParams.page) || 1
  const keyword = searchParams.keyword || ''

  let ershouPage = { list: [] as any[], total: 0, page, pageSize: 12 }
  try {
    ershouPage = await listErshous({ page, pageSize: 12, keyword })
  } catch {
    // 后端不可达或模块未启用，渲染空状态
  }

  const totalPages = Math.ceil(ershouPage.total / ershouPage.pageSize) || 1

  return (
    <div className="container py-6">
      <nav className="text-sm text-gray-500 mb-4">
        <Link href="/" className="hover:text-brand-600">首页</Link>
        <span className="mx-2">/</span>
        <span>二手交易</span>
      </nav>

      <h1 className="text-2xl font-bold mb-4">
        二手交易{keyword ? ` · 搜索：${keyword}` : ''}
      </h1>

      {ershouPage.list.length === 0 ? (
        <div className="bg-white rounded-lg p-12 text-center text-gray-500 border border-gray-200">
          暂无二手物品
          <div className="mt-2 text-xs text-gray-400">
            （后端 /ershou 接口未就绪时也会展示此状态）
          </div>
        </div>
      ) : (
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">
          {ershouPage.list.map((item) => (
            <Link
              key={item.id}
              href={`/ershou/${item.id}`}
              className="block bg-white rounded-lg overflow-hidden border border-gray-200 hover:shadow-md transition-shadow"
            >
              {item.cover_image && (
                // eslint-disable-next-line @next/next/no-img-element
                <img
                  src={item.cover_image}
                  alt={item.title}
                  className="w-full h-48 object-cover"
                />
              )}
              <div className="p-4">
                <h3 className="font-bold text-base mb-2 line-clamp-2 hover:text-brand-600">
                  {item.title}
                </h3>
                <div className="flex items-center justify-between mb-2">
                  <span className="text-brand-600 font-bold text-lg">
                    ¥{item.price}
                  </span>
                  {item.original_price && (
                    <span className="text-xs text-gray-400 line-through">
                      ¥{item.original_price}
                    </span>
                  )}
                </div>
                <div className="flex items-center justify-between text-xs text-gray-500">
                  <span>{item.region_name || '五常'}</span>
                  <span>👁 {item.view_count || 0}</span>
                </div>
              </div>
            </Link>
          ))}
        </div>
      )}

      {totalPages > 1 && (
        <div className="flex items-center justify-center gap-2 mt-8">
          {page > 1 && (
            <Link
              href={`/ershou?page=${page - 1}${keyword ? `&keyword=${encodeURIComponent(keyword)}` : ''}`}
              className="px-3 py-1.5 text-sm border rounded hover:bg-gray-50"
            >
              上一页
            </Link>
          )}
          <span className="text-sm text-gray-600">
            第 {page} / {totalPages} 页（共 {ershouPage.total} 条）
          </span>
          {page < totalPages && (
            <Link
              href={`/ershou?page=${page + 1}${keyword ? `&keyword=${encodeURIComponent(keyword)}` : ''}`}
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
