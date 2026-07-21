import Link from 'next/link'
import { notFound } from 'next/navigation'
import type { Metadata } from 'next'
import {
  getDh114Detail,
  listDh114s,
  listDh114Groupbuys,
  listDh114Coupons,
} from '@/lib/api'

export const revalidate = 0

interface PageProps {
  params: { id: string }
}

export async function generateMetadata({ params }: PageProps): Promise<Metadata> {
  const id = Number(params.id)
  if (Number.isNaN(id)) return { title: '商户详情' }
  try {
    const item = await getDh114Detail(id)
    return {
      title: `${item.name} - 同城114`,
      description: item.description?.slice(0, 120) || '武昌同城114商户详情',
      keywords: [item.name, '武昌商家', '武昌黄页', item.category_name || ''],
      openGraph: {
        title: item.name,
        description: item.description?.slice(0, 120),
        type: 'article',
        images: item.cover_image ? [{ url: item.cover_image }] : undefined,
      },
    }
  } catch {
    return { title: '商户详情' }
  }
}

export default async function Dh114DetailPage({ params }: PageProps) {
  const id = Number(params.id)
  if (Number.isNaN(id)) notFound()

  let item
  let related: any[] = []
  let groupbuys: any[] = []
  let coupons: any[] = []
  try {
    item = await getDh114Detail(id)
  } catch {
    notFound()
  }

  try {
    const rel = await listDh114s({ page: 1, pageSize: 6, categoryId: item.category_id })
    related = rel.list.filter((r) => r.id !== item.id).slice(0, 5)
  } catch {
    // 后端不可达
  }
  try {
    const gb = await listDh114Groupbuys({ dh114Id: id, page: 1, pageSize: 10 })
    groupbuys = gb.list || []
  } catch {
    // 团购接口不可达
  }
  try {
    const cp = await listDh114Coupons({ dh114Id: id, page: 1, pageSize: 10 })
    coupons = cp.list || []
  } catch {
    // 优惠券接口不可达
  }

  const statusText = item.status === 1 ? '营业中' : item.status === 2 ? '休息中' : '已关闭'

  return (
    <div className="container py-6">
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* 主信息 */}
        <div className="lg:col-span-2 space-y-4">
          <div className="bg-white rounded-lg p-6 border border-gray-200">
            <nav className="text-sm text-gray-500 mb-4">
              <Link href="/" className="hover:text-brand-600">首页</Link>
              <span className="mx-2">/</span>
              <Link href="/dh114" className="hover:text-brand-600">同城114</Link>
              <span className="mx-2">/</span>
              <span>{item.name}</span>
            </nav>

            <div className="flex items-start gap-4 mb-6">
              <div className="w-32 h-32 rounded overflow-hidden bg-gray-100 flex-shrink-0">
                {item.cover_image || item.logo ? (
                  // eslint-disable-next-line @next/next/no-img-element
                  <img src={item.cover_image || item.logo} alt={item.name} className="w-full h-full object-cover" />
                ) : (
                  <div className="w-full h-full flex items-center justify-center text-brand-400 text-3xl font-bold">
                    {item.name?.charAt(0) || '商'}
                  </div>
                )}
              </div>
              <div className="flex-1">
                <div className="flex items-center gap-2 mb-2">
                  <h1 className="text-2xl font-bold">{item.name}</h1>
                  <span className={`px-2 py-0.5 text-xs rounded ${
                    item.status === 1 ? 'bg-green-100 text-green-700' : 'bg-gray-100 text-gray-500'
                  }`}>{statusText}</span>
                </div>
                <div className="grid grid-cols-2 gap-y-1 gap-x-4 text-sm text-gray-600">
                  {item.category_name && <div>分类：<span className="text-gray-900">{item.category_name}</span></div>}
                  {item.region_name && <div>地区：<span className="text-gray-900">{item.region_name}</span></div>}
                  {item.business_hours && <div>营业：<span className="text-gray-900">{item.business_hours}</span></div>}
                  {item.phone && <div>电话：<span className="text-gray-900">{item.phone}</span></div>}
                  {item.mobile && <div>手机：<span className="text-gray-900">{item.mobile}</span></div>}
                  {item.address && <div className="col-span-2">地址：<span className="text-gray-900">{item.address}</span></div>}
                </div>
              </div>
            </div>

            {/* 关键指标 */}
            <div className="grid grid-cols-2 sm:grid-cols-4 gap-4 py-4 border-y border-gray-100">
              <div>
                <div className="text-xs text-gray-500 mb-1">评分</div>
                <div className="font-bold text-brand-600">{item.rating ? `⭐ ${item.rating.toFixed(1)}` : '暂无'}</div>
              </div>
              <div>
                <div className="text-xs text-gray-500 mb-1">人均</div>
                <div className="font-bold">{item.avg_price ? `¥${item.avg_price}` : '-'}</div>
              </div>
              <div>
                <div className="text-xs text-gray-500 mb-1">评论数</div>
                <div className="font-bold">{item.review_count || 0}</div>
              </div>
              <div>
                <div className="text-xs text-gray-500 mb-1">浏览量</div>
                <div className="font-bold">👁 {item.view_count || 0}</div>
              </div>
            </div>

            {item.description && (
              <div className="mt-4">
                <h3 className="font-bold mb-2">商家简介</h3>
                <p className="text-gray-700 whitespace-pre-line">{item.description}</p>
              </div>
            )}
          </div>

          {/* 团购 */}
          {groupbuys.length > 0 && (
            <div className="bg-white rounded-lg p-6 border border-gray-200">
              <h3 className="font-bold mb-3">本店团购（{groupbuys.length}）</h3>
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
                {groupbuys.map((gb) => (
                  <Link
                    key={gb.id}
                    href={`/groupbuy/${gb.id}`}
                    className="flex items-center gap-3 p-2 border border-gray-200 rounded hover:shadow-sm"
                  >
                    <div className="w-16 h-16 bg-gray-100 rounded flex-shrink-0 overflow-hidden">
                      {gb.cover_image ? (
                        // eslint-disable-next-line @next/next/no-img-element
                        <img src={gb.cover_image} alt={gb.title} className="w-full h-full object-cover" />
                      ) : null}
                    </div>
                    <div className="flex-1 min-w-0">
                      <div className="font-medium text-sm truncate">{gb.title}</div>
                      <div className="flex items-baseline gap-1 mt-1">
                        <span className="text-brand-600 font-bold">¥{gb.price}</span>
                        {gb.original_price && (
                          <span className="text-xs text-gray-400 line-through">¥{gb.original_price}</span>
                        )}
                      </div>
                    </div>
                  </Link>
                ))}
              </div>
            </div>
          )}

          {/* 优惠券 */}
          {coupons.length > 0 && (
            <div className="bg-white rounded-lg p-6 border border-gray-200">
              <h3 className="font-bold mb-3">本店优惠券（{coupons.length}）</h3>
              <div className="space-y-2">
                {coupons.map((cp) => (
                  <div key={cp.id} className="flex items-center justify-between p-3 bg-orange-50 border border-orange-100 rounded">
                    <div>
                      <div className="font-medium text-sm">{cp.title}</div>
                      <div className="text-xs text-gray-500 mt-0.5">
                        {cp.type === 1 ? `满${cp.threshold}减${cp.amount}` :
                         cp.type === 2 ? `${cp.discount}折` :
                         cp.type === 3 ? `代金券 ¥${cp.amount}` : '优惠券'}
                      </div>
                    </div>
                    <div className="text-xs text-gray-500">
                      剩余 {cp.total_count - cp.received_count}/{cp.total_count}
                    </div>
                  </div>
                ))}
              </div>
            </div>
          )}
        </div>

        {/* 侧边栏 */}
        <aside className="space-y-4">
          <div className="bg-white rounded-lg p-4 border border-gray-200">
            <h3 className="font-bold mb-3">同分类商户</h3>
            {related.length === 0 ? (
              <p className="text-sm text-gray-500">暂无相关商户</p>
            ) : (
              <div className="space-y-3">
                {related.map((r) => (
                  <Link
                    key={r.id}
                    href={`/dh114/${r.id}`}
                    className="flex items-center gap-3 p-2 rounded hover:bg-gray-50"
                  >
                    <div className="w-12 h-12 bg-gray-100 rounded flex-shrink-0 overflow-hidden">
                      {r.cover_image ? (
                        // eslint-disable-next-line @next/next/no-img-element
                        <img src={r.cover_image} alt={r.name} className="w-full h-full object-cover" />
                      ) : (
                        <div className="w-full h-full flex items-center justify-center text-brand-400 font-bold">
                          {r.name?.charAt(0) || '商'}
                        </div>
                      )}
                    </div>
                    <div className="flex-1 min-w-0">
                      <div className="font-medium text-sm truncate">{r.name}</div>
                      <div className="text-xs text-gray-500 flex items-center justify-between">
                        <span>{r.category_name || ''}</span>
                        {r.rating && <span className="text-brand-600">⭐ {r.rating.toFixed(1)}</span>}
                      </div>
                    </div>
                  </Link>
                ))}
              </div>
            )}
          </div>
        </aside>
      </div>
    </div>
  )
}
