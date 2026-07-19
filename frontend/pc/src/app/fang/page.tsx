import Link from 'next/link'
import type { Metadata } from 'next'
import { listFangs } from '@/lib/api'

export const revalidate = 60

export const metadata: Metadata = {
  title: '房屋租售',
  description:
    '五常同城房屋租售频道，提供本地房屋出租、出售、合租、转让等信息。',
  keywords: ['五常房产', '五常租房', '五常二手房', '五常房屋出租'],
  openGraph: {
    title: '五常同城房屋租售',
    description: '本地房屋出租、出售、合租、转让信息',
    type: 'website',
  },
}

interface SearchParams {
  page?: string
  keyword?: string
}

export default async function FangListPage({
  searchParams,
}: {
  searchParams: SearchParams
}) {
  const page = Number(searchParams.page) || 1
  const keyword = searchParams.keyword || ''

  let fangPage = { list: [] as any[], total: 0, page, pageSize: 12 }
  try {
    fangPage = await listFangs({ page, pageSize: 12, keyword })
  } catch {
    // 后端 /fang 接口未就绪，渲染空状态
  }

  const totalPages = Math.ceil(fangPage.total / fangPage.pageSize) || 1

  return (
    <div className="container py-6">
      <nav className="text-sm text-gray-500 mb-4">
        <Link href="/" className="hover:text-brand-600">首页</Link>
        <span className="mx-2">/</span>
        <span>房屋租售</span>
      </nav>

      <h1 className="text-2xl font-bold mb-4">
        房屋租售{keyword ? ` · 搜索：${keyword}` : ''}
      </h1>

      {fangPage.list.length === 0 ? (
        <div className="bg-white rounded-lg p-12 text-center text-gray-500 border border-gray-200">
          暂无房源信息
          <div className="mt-2 text-xs text-gray-400">
            （后端 /fang 接口未就绪时也会展示此状态）
          </div>
        </div>
      ) : (
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
          {fangPage.list.map((item) => (
            <Link
              key={item.id}
              href={`/fang/${item.id}`}
              className="block bg-white rounded-lg p-4 border border-gray-200 hover:shadow-md transition-shadow"
            >
              <div className="flex items-center justify-between mb-2">
                <h3 className="font-bold text-base line-clamp-1 hover:text-brand-600">
                  {item.title}
                </h3>
                <span className="text-xs px-2 py-0.5 bg-brand-50 text-brand-600 rounded whitespace-nowrap">
                  {item.type}
                </span>
              </div>
              <div className="text-brand-600 font-bold text-lg mb-2">
                {item.price}
              </div>
              <div className="flex items-center gap-3 text-xs text-gray-500">
                {item.layout && <span>{item.layout}</span>}
                {item.area && <span>{item.area}</span>}
                {item.location && <span>📍 {item.location}</span>}
              </div>
            </Link>
          ))}
        </div>
      )}

      {totalPages > 1 && (
        <div className="flex items-center justify-center gap-2 mt-8">
          {page > 1 && (
            <Link
              href={`/fang?page=${page - 1}${keyword ? `&keyword=${encodeURIComponent(keyword)}` : ''}`}
              className="px-3 py-1.5 text-sm border rounded hover:bg-gray-50"
            >
              上一页
            </Link>
          )}
          <span className="text-sm text-gray-600">
            第 {page} / {totalPages} 页（共 {fangPage.total} 条）
          </span>
          {page < totalPages && (
            <Link
              href={`/fang?page=${page + 1}${keyword ? `&keyword=${encodeURIComponent(keyword)}` : ''}`}
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
