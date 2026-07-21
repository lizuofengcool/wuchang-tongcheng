import Link from 'next/link'
import type { Metadata } from 'next'
import { listLinggongs } from '@/lib/api'
import LinggongCard from '@/components/LinggongCard'

export const revalidate = 60

export const metadata: Metadata = {
  title: '零工兼职',
  description:
    '武昌同城零工兼职频道，提供本地零工、兼职、全职招聘信息，包括日结工、小时工、临时工、暑期工等。',
  keywords: ['武昌零工', '武昌兼职', '日结工', '小时工', '武昌同城兼职'],
  openGraph: {
    title: '武昌同城零工兼职',
    description: '本地零工、兼职、全职招聘信息',
    type: 'website',
  },
}

interface SearchParams {
  page?: string
  type?: string
  categoryId?: string
  keyword?: string
}

export default async function LinggongListPage({
  searchParams,
}: {
  searchParams: SearchParams
}) {
  const page = Number(searchParams.page) || 1
  const type = searchParams.type ? Number(searchParams.type) : undefined
  const categoryId = searchParams.categoryId ? Number(searchParams.categoryId) : undefined
  const keyword = searchParams.keyword || ''

  let linggongPage = { list: [] as any[], total: 0, page, pageSize: 12 }
  try {
    linggongPage = await listLinggongs({ page, pageSize: 12, type, categoryId, keyword })
  } catch {
    // 后端 /linggong 接口未就绪
  }

  const totalPages = Math.ceil(linggongPage.total / linggongPage.pageSize) || 1

  const buildPageHref = (p: number) => {
    const q = new URLSearchParams()
    if (p !== 1) q.set('page', String(p))
    if (type) q.set('type', String(type))
    if (categoryId) q.set('categoryId', String(categoryId))
    if (keyword) q.set('keyword', keyword)
    const s = q.toString()
    return s ? `/linggong?${s}` : '/linggong'
  }

  const typeFilters = [
    { value: 0, label: '不限' },
    { value: 1, label: '零工' },
    { value: 2, label: '兼职' },
    { value: 3, label: '全职' },
  ]

  return (
    <div className="container py-6">
      <nav className="text-sm text-gray-500 mb-4">
        <Link href="/" className="hover:text-brand-600">首页</Link>
        <span className="mx-2">/</span>
        <span>零工兼职</span>
      </nav>

      <h1 className="text-2xl font-bold mb-4">
        零工兼职{keyword ? ` · 搜索：${keyword}` : ''}
      </h1>

      {/* 筛选区 */}
      <div className="bg-white rounded-lg p-4 mb-6 border border-gray-200">
        <div className="flex flex-wrap items-center gap-4 text-sm">
          <div className="flex items-center gap-2">
            <span className="text-gray-500">类型：</span>
            {typeFilters.map((tf) => {
              const href = `${buildPageHref(1)}${buildPageHref(1).includes('?') ? '&' : '?'}${tf.value ? `type=${tf.value}` : 'type='}`
              return (
                <Link
                  key={tf.value}
                  href={href}
                  className={`px-2 py-1 rounded ${(type ?? 0) === tf.value ? 'bg-brand-600 text-white' : 'hover:bg-gray-100'}`}
                >
                  {tf.label}
                </Link>
              )
            })}
          </div>

          <form className="flex items-center gap-2 ml-auto" method="get" action="/linggong">
            <input
              type="text"
              name="keyword"
              defaultValue={keyword}
              placeholder="搜索岗位/雇主…"
              className="w-48 px-3 py-1.5 text-sm border border-gray-300 rounded focus:outline-none focus:border-brand-500"
            />
            <button
              type="submit"
              className="px-3 py-1.5 text-sm bg-brand-600 text-white rounded hover:bg-brand-700"
            >
              搜索
            </button>
          </form>
        </div>
      </div>

      {linggongPage.list.length === 0 ? (
        <div className="bg-white rounded-lg p-12 text-center text-gray-500 border border-gray-200">
          暂无零工兼职信息
          <div className="mt-2 text-xs text-gray-400">
            （后端 /linggong 接口未就绪时也会展示此状态）
          </div>
        </div>
      ) : (
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
          {linggongPage.list.map((item) => (
            <LinggongCard key={item.id} linggong={item} />
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
            第 {page} / {totalPages} 页（共 {linggongPage.total} 条）
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
