import Link from 'next/link'
import type { Metadata } from 'next'
import { listGroupBuys, listGroupBuyCoupons } from '@/lib/api'
import GroupBuyCard from '@/components/GroupBuyCard'

export const revalidate = 60

export const metadata: Metadata = {
  title: '同城团购',
  description:
    '武昌同城团购频道，提供本地团购优惠、折扣商品、限时抢购、超值套餐，省钱又划算。',
  keywords: ['武昌团购', '同城团购', '团购优惠', '武昌折扣'],
  openGraph: {
    title: '武昌同城团购',
    description: '本地团购优惠、折扣商品、限时抢购',
    type: 'website',
  },
}

interface SearchParams {
  page?: string
  keyword?: string
}

export default async function GroupBuyListPage({
  searchParams,
}: {
  searchParams: SearchParams
}) {
  const page = Number(searchParams.page) || 1
  const keyword = searchParams.keyword || ''

  let gbPage = { list: [] as any[], total: 0, page, pageSize: 12 }
  let coupons: any[] = []
  try {
    gbPage = await listGroupBuys({ page, pageSize: 12, keyword })
  } catch {
    // 后端 /groupbuy/list 接口未就绪
  }
  try {
    const cp = await listGroupBuyCoupons({ page: 1, pageSize: 8 })
    coupons = cp.list || []
  } catch {
    // 优惠券接口不可达
  }

  const totalPages = Math.ceil(gbPage.total / gbPage.pageSize) || 1

  const buildPageHref = (p: number) => {
    const q = new URLSearchParams()
    if (p !== 1) q.set('page', String(p))
    if (keyword) q.set('keyword', keyword)
    const s = q.toString()
    return s ? `/groupbuy?${s}` : '/groupbuy'
  }

  return (
    <div className="container py-6">
      <nav className="text-sm text-gray-500 mb-4">
        <Link href="/" className="hover:text-brand-600">首页</Link>
        <span className="mx-2">/</span>
        <span>同城团购</span>
      </nav>

      <h1 className="text-2xl font-bold mb-4">
        同城团购{keyword ? ` · 搜索：${keyword}` : ''}
      </h1>

      {/* 搜索 */}
      <div className="bg-white rounded-lg p-4 mb-6 border border-gray-200">
        <form className="flex items-center gap-2" method="get" action="/groupbuy">
          <input
            type="text"
            name="keyword"
            defaultValue={keyword}
            placeholder="搜索团购商品…"
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

      {/* 优惠券 */}
      {coupons.length > 0 && (
        <div className="bg-white rounded-lg p-4 mb-6 border border-gray-200">
          <h3 className="font-bold mb-3 text-sm">热门优惠券</h3>
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-3">
            {coupons.map((cp) => (
              <div
                key={cp.id}
                className="p-3 bg-gradient-to-br from-orange-50 to-brand-100 border border-orange-100 rounded"
              >
                <div className="font-medium text-sm truncate">{cp.title}</div>
                <div className="text-xs text-gray-600 mt-1">
                  {cp.type === 1 ? `满${cp.threshold || 0}减${cp.amount || 0}` :
                   cp.type === 2 ? `${cp.discount || 0}折` :
                   cp.type === 3 ? `代金券 ¥${cp.amount || 0}` : '优惠券'}
                </div>
                <div className="text-xs text-gray-400 mt-1">
                  剩余 {cp.total_count - cp.received_count}/{cp.total_count}
                </div>
              </div>
            ))}
          </div>
        </div>
      )}

      {gbPage.list.length === 0 ? (
        <div className="bg-white rounded-lg p-12 text-center text-gray-500 border border-gray-200">
          暂无团购商品
          <div className="mt-2 text-xs text-gray-400">
            （后端 /groupbuy/list 接口未就绪时也会展示此状态）
          </div>
        </div>
      ) : (
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">
          {gbPage.list.map((item) => (
            <GroupBuyCard key={item.id} groupbuy={item} />
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
            第 {page} / {totalPages} 页（共 {gbPage.total} 个团购）
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
