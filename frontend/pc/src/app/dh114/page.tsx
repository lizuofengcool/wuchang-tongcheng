import Link from 'next/link'
import type { Metadata } from 'next'
import { listDh114s, listDh114Categories } from '@/lib/api'
import Dh114Card from '@/components/Dh114Card'

export const revalidate = 60

export const metadata: Metadata = {
  title: '同城114',
  description:
    '武昌同城114黄页频道，提供本地商户、商家、企业信息，包括餐饮、零售、服务、医疗、教育等分类。',
  keywords: ['武昌黄页', '武昌114', '本地商户', '武昌商家', '武昌企业'],
  openGraph: {
    title: '武昌同城114黄页',
    description: '本地商户、商家、企业信息查询',
    type: 'website',
  },
}

interface SearchParams {
  page?: string
  categoryId?: string
  keyword?: string
}

export default async function Dh114ListPage({
  searchParams,
}: {
  searchParams: SearchParams
}) {
  const page = Number(searchParams.page) || 1
  const categoryId = searchParams.categoryId ? Number(searchParams.categoryId) : undefined
  const keyword = searchParams.keyword || ''

  let dh114Page = { list: [] as any[], total: 0, page, pageSize: 12 }
  let categories: any[] = []
  try {
    dh114Page = await listDh114s({ page, pageSize: 12, categoryId, keyword })
  } catch {
    // 后端 /dh114 接口未就绪
  }
  try {
    categories = await listDh114Categories()
  } catch {
    // 分类接口不可达
  }

  const totalPages = Math.ceil(dh114Page.total / dh114Page.pageSize) || 1

  const buildPageHref = (p: number) => {
    const q = new URLSearchParams()
    if (p !== 1) q.set('page', String(p))
    if (categoryId) q.set('categoryId', String(categoryId))
    if (keyword) q.set('keyword', keyword)
    const s = q.toString()
    return s ? `/dh114?${s}` : '/dh114'
  }

  return (
    <div className="container py-6">
      <nav className="text-sm text-gray-500 mb-4">
        <Link href="/" className="hover:text-brand-600">首页</Link>
        <span className="mx-2">/</span>
        <span>同城114</span>
      </nav>

      <h1 className="text-2xl font-bold mb-4">
        同城114{keyword ? ` · 搜索：${keyword}` : ''}
      </h1>

      {/* 分类导航 */}
      {categories.length > 0 && (
        <div className="bg-white rounded-lg p-4 mb-6 border border-gray-200">
          <h3 className="font-bold mb-3 text-sm">分类导航</h3>
          <div className="flex flex-wrap gap-2 text-sm">
            <Link
              href="/dh114"
              className={`px-3 py-1.5 rounded ${!categoryId ? 'bg-brand-600 text-white' : 'bg-gray-100 hover:bg-brand-50 hover:text-brand-600'}`}
            >
              全部
            </Link>
            {categories.map((c) => (
              <Link
                key={c.id}
                href={`/dh114?categoryId=${c.id}`}
                className={`px-3 py-1.5 rounded ${categoryId === c.id ? 'bg-brand-600 text-white' : 'bg-gray-100 hover:bg-brand-50 hover:text-brand-600'}`}
              >
                {c.name}
              </Link>
            ))}
          </div>
        </div>
      )}

      {/* 搜索区 */}
      <div className="bg-white rounded-lg p-4 mb-6 border border-gray-200">
        <form className="flex items-center gap-2" method="get" action="/dh114">
          <input
            type="text"
            name="keyword"
            defaultValue={keyword}
            placeholder="搜索商家名称/类别…"
            className="flex-1 px-3 py-1.5 text-sm border border-gray-300 rounded focus:outline-none focus:border-brand-500"
          />
          <button
            type="submit"
            className="px-4 py-1.5 text-sm bg-brand-600 text-white rounded hover:bg-brand-700"
          >
            搜索
          </button>
        </form>
      </div>

      {dh114Page.list.length === 0 ? (
        <div className="bg-white rounded-lg p-12 text-center text-gray-500 border border-gray-200">
          暂无商户信息
          <div className="mt-2 text-xs text-gray-400">
            （后端 /dh114 接口未就绪时也会展示此状态）
          </div>
        </div>
      ) : (
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">
          {dh114Page.list.map((item) => (
            <Dh114Card key={item.id} dh114={item} />
          ))}
        </div>
      )}

      {totalPages > 1 && (
        <div className="flex items-center justify-center gap-2 mt-8">
          {page > 1 && (
            <Link
              href={buildPageHref(page - 1)}
              className="px-3 py-1.5 text-sm border rounded hover:bg-gray-50"
            >
              上一页
            </Link>
          )}
          <span className="text-sm text-gray-600">
            第 {page} / {totalPages} 页（共 {dh114Page.total} 家商户）
          </span>
          {page < totalPages && (
            <Link
              href={buildPageHref(page + 1)}
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
