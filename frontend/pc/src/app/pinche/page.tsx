import Link from 'next/link'
import type { Metadata } from 'next'
import { listPinches, listPincheRoutes } from '@/lib/api'
import PincheCard from '@/components/PincheCard'

export const revalidate = 60

export const metadata: Metadata = {
  title: '拼车出行',
  description:
    '武昌同城拼车出行频道，提供本地拼车信息、顺风车、长途拼车、上下班拼车等出行服务。',
  keywords: ['武昌拼车', '顺风车', '同城拼车', '武昌出行'],
  openGraph: {
    title: '武昌同城拼车出行',
    description: '本地拼车、顺风车、长途拼车信息',
    type: 'website',
  },
}

interface SearchParams {
  page?: string
  startCity?: string
  endCity?: string
  date?: string
  keyword?: string
}

export default async function PincheListPage({
  searchParams,
}: {
  searchParams: SearchParams
}) {
  const page = Number(searchParams.page) || 1
  const startCity = searchParams.startCity || ''
  const endCity = searchParams.endCity || ''
  const date = searchParams.date || ''
  const keyword = searchParams.keyword || ''

  let pinchePage = { list: [] as any[], total: 0, page, pageSize: 12 }
  let routes: any[] = []
  try {
    [pinchePage] = await Promise.all([
      listPinches({ page, pageSize: 12, startCity, endCity, date, keyword }),
    ])
  } catch {
    // 后端不可达
  }
  try {
    const r = await listPincheRoutes({ page: 1, pageSize: 10 })
    routes = r.list || []
  } catch {
    // 路线接口不可达
  }

  const totalPages = Math.ceil(pinchePage.total / pinchePage.pageSize) || 1

  const buildPageHref = (p: number) => {
    const q = new URLSearchParams()
    if (p !== 1) q.set('page', String(p))
    if (startCity) q.set('startCity', startCity)
    if (endCity) q.set('endCity', endCity)
    if (date) q.set('date', date)
    if (keyword) q.set('keyword', keyword)
    const s = q.toString()
    return s ? `/pinche?${s}` : '/pinche'
  }

  return (
    <div className="container py-6">
      <nav className="text-sm text-gray-500 mb-4">
        <Link href="/" className="hover:text-brand-600">首页</Link>
        <span className="mx-2">/</span>
        <span>拼车出行</span>
      </nav>

      <h1 className="text-2xl font-bold mb-4">
        拼车出行{keyword ? ` · 搜索：${keyword}` : ''}
      </h1>

      {/* 筛选区 */}
      <div className="bg-white rounded-lg p-4 mb-6 border border-gray-200">
        <form className="flex flex-wrap items-center gap-3 text-sm" method="get" action="/pinche">
          <div className="flex items-center gap-1">
            <input
              type="text"
              name="startCity"
              defaultValue={startCity}
              placeholder="出发城市"
              className="w-32 px-3 py-1.5 border border-gray-300 rounded focus:outline-none focus:border-brand-500"
            />
            <span className="text-gray-400">→</span>
            <input
              type="text"
              name="endCity"
              defaultValue={endCity}
              placeholder="到达城市"
              className="w-32 px-3 py-1.5 border border-gray-300 rounded focus:outline-none focus:border-brand-500"
            />
          </div>
          <input
            type="date"
            name="date"
            defaultValue={date}
            className="px-3 py-1.5 border border-gray-300 rounded focus:outline-none focus:border-brand-500"
          />
          <input
            type="text"
            name="keyword"
            defaultValue={keyword}
            placeholder="关键词…"
            className="flex-1 min-w-[12rem] px-3 py-1.5 border border-gray-300 rounded focus:outline-none focus:border-brand-500"
          />
          <button
            type="submit"
            className="px-4 py-1.5 bg-brand-600 text-white rounded hover:bg-brand-700"
          >
            搜索
          </button>
        </form>
      </div>

      {/* 热门路线 */}
      {routes.length > 0 && (
        <div className="bg-white rounded-lg p-4 mb-6 border border-gray-200">
          <h3 className="font-bold mb-3">热门路线</h3>
          <div className="flex flex-wrap gap-2 text-sm">
            {routes.map((r) => (
              <Link
                key={r.id}
                href={`/pinche?startCity=${encodeURIComponent(r.start_location)}&endCity=${encodeURIComponent(r.end_location)}`}
                className="px-3 py-1.5 bg-gray-100 rounded hover:bg-brand-50 hover:text-brand-600"
              >
                {r.start_location} → {r.end_location}
                {r.use_count ? <span className="text-xs text-gray-400 ml-1">({r.use_count})</span> : null}
              </Link>
            ))}
          </div>
        </div>
      )}

      {pinchePage.list.length === 0 ? (
        <div className="bg-white rounded-lg p-12 text-center text-gray-500 border border-gray-200">
          暂无拼车信息
          <div className="mt-2 text-xs text-gray-400">
            （后端 /pinche 接口未就绪时也会展示此状态）
          </div>
        </div>
      ) : (
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
          {pinchePage.list.map((item) => (
            <PincheCard key={item.id} pinche={item} />
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
            第 {page} / {totalPages} 页（共 {pinchePage.total} 条）
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
