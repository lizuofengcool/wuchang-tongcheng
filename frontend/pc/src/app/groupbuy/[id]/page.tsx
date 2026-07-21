import Link from 'next/link'
import { notFound } from 'next/navigation'
import type { Metadata } from 'next'
import { getGroupBuyDetail, listGroupBuys, listGroupBuyCoupons } from '@/lib/api'

export const revalidate = 0

interface PageProps {
  params: { id: string }
}

export async function generateMetadata({ params }: PageProps): Promise<Metadata> {
  const id = Number(params.id)
  if (Number.isNaN(id)) return { title: '团购详情' }
  try {
    const item = await getGroupBuyDetail(id)
    return {
      title: `${item.title} - 同城团购`,
      description: item.subtitle?.slice(0, 120) || item.description?.slice(0, 120) || '武昌同城团购详情',
      keywords: [item.title, '武昌团购', '同城团购'],
      openGraph: {
        title: item.title,
        description: item.subtitle?.slice(0, 120),
        type: 'article',
        images: item.cover_image ? [{ url: item.cover_image }] : undefined,
      },
    }
  } catch {
    return { title: '团购详情' }
  }
}

export default async function GroupBuyDetailPage({ params }: PageProps) {
  const id = Number(params.id)
  if (Number.isNaN(id)) notFound()

  let item
  let related: any[] = []
  let coupons: any[] = []
  try {
    item = await getGroupBuyDetail(id)
  } catch {
    notFound()
  }

  try {
    const rel = await listGroupBuys({ page: 1, pageSize: 6, shopId: item.shop_id })
    related = rel.list.filter((r) => r.id !== item.id).slice(0, 5)
  } catch {
    // 后端不可达
  }
  try {
    const cp = await listGroupBuyCoupons({ groupbuyId: id, page: 1, pageSize: 10 })
    coupons = cp.list || []
  } catch {
    // 优惠券接口不可达
  }

  const discount =
    item.original_price && item.original_price > 0
      ? Math.round((item.price / item.original_price) * 10)
      : 0
  const saved = item.original_price && item.original_price > item.price
    ? item.original_price - item.price
    : 0

  return (
    <div className="container py-6">
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* 主信息 */}
        <div className="lg:col-span-2 bg-white rounded-lg p-6 border border-gray-200">
          <nav className="text-sm text-gray-500 mb-4">
            <Link href="/" className="hover:text-brand-600">首页</Link>
            <span className="mx-2">/</span>
            <Link href="/groupbuy" className="hover:text-brand-600">同城团购</Link>
            <span className="mx-2">/</span>
            <span>{item.title}</span>
          </nav>

          {/* 团购图 + 信息 */}
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-6 mb-6">
            <div className="relative bg-gray-50 rounded-lg overflow-hidden">
              {item.cover_image ? (
                // eslint-disable-next-line @next/next/no-img-element
                <img src={item.cover_image} alt={item.title} className="w-full h-80 object-cover" />
              ) : (
                <div className="w-full h-80 flex items-center justify-center text-gray-400 text-sm">暂无图片</div>
              )}
              {discount > 0 && (
                <span className="absolute top-2 right-2 px-2 py-1 text-sm bg-brand-600 text-white rounded">
                  {discount}折
                </span>
              )}
            </div>
            <div>
              <h1 className="text-xl font-bold mb-2">{item.title}</h1>
              {item.subtitle && (
                <p className="text-sm text-gray-500 mb-3">{item.subtitle}</p>
              )}

              {/* 价格 */}
              <div className="bg-gradient-to-br from-orange-50 to-brand-100 p-4 rounded mb-4">
                <div className="flex items-baseline gap-2">
                  <span className="text-brand-600 font-bold text-3xl">¥{item.price}</span>
                  {item.original_price && item.original_price > item.price && (
                    <span className="text-sm text-gray-400 line-through">¥{item.original_price}</span>
                  )}
                </div>
                {saved > 0 && (
                  <div className="text-xs text-brand-700 mt-1">立省 ¥{saved}</div>
                )}
              </div>

              {/* 关键信息 */}
              <div className="grid grid-cols-2 gap-y-2 gap-x-4 text-sm text-gray-600 mb-4">
                <div>销量：<span className="text-gray-900">{item.sales_count || 0}</span></div>
                <div>库存：<span className="text-gray-900">{item.stock}</span></div>
                {item.limit_per_user && <div>限购：<span className="text-gray-900">{item.limit_per_user} 件/人</span></div>}
                {item.shop_name && <div className="col-span-2">商家：<span className="text-gray-900">{item.shop_name}</span></div>}
              </div>

              {/* 购买按钮 */}
              <div className="flex items-center gap-3">
                <button
                  type="button"
                  disabled
                  className="px-8 py-2.5 bg-brand-600 text-white rounded text-lg hover:bg-brand-700 disabled:opacity-50"
                >
                  立即抢购
                </button>
                <span className="text-xs text-gray-400">请登录后操作</span>
              </div>

              {/* 时间 */}
              {(item.start_time || item.end_time) && (
                <div className="mt-3 text-xs text-gray-500">
                  {item.start_time && <span>开始：{new Date(item.start_time).toLocaleString('zh-CN')}</span>}
                  {item.end_time && <span className="ml-3">结束：{new Date(item.end_time).toLocaleString('zh-CN')}</span>}
                </div>
              )}
            </div>
          </div>

          {/* 多图 */}
          {item.images && item.images.length > 0 && (
            <div className="mb-6">
              <h3 className="font-bold mb-2">商品图片</h3>
              <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 gap-2">
                {item.images.map((img, i) => (
                  // eslint-disable-next-line @next/next/no-img-element
                  <img
                    key={i}
                    src={img}
                    alt={`${item.title} ${i + 1}`}
                    className="w-full h-32 object-cover rounded"
                  />
                ))}
              </div>
            </div>
          )}

          {/* 详情 */}
          {item.description && (
            <div className="mb-6">
              <h3 className="font-bold mb-2">团购详情</h3>
              <div className="text-gray-700 whitespace-pre-line text-sm leading-relaxed">
                {item.description}
              </div>
            </div>
          )}

          {/* 优惠券 */}
          {coupons.length > 0 && (
            <div className="mb-6">
              <h3 className="font-bold mb-2">可用优惠券</h3>
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
            <h3 className="font-bold mb-3">相关团购</h3>
            {related.length === 0 ? (
              <p className="text-sm text-gray-500">暂无相关团购</p>
            ) : (
              <div className="space-y-3">
                {related.map((r) => (
                  <Link
                    key={r.id}
                    href={`/groupbuy/${r.id}`}
                    className="flex items-center gap-3 hover:bg-gray-50 p-2 rounded"
                  >
                    <div className="w-12 h-12 bg-gray-100 rounded flex-shrink-0 overflow-hidden">
                      {r.cover_image ? (
                        // eslint-disable-next-line @next/next/no-img-element
                        <img src={r.cover_image} alt={r.title} className="w-full h-full object-cover" />
                      ) : null}
                    </div>
                    <div className="flex-1 min-w-0">
                      <div className="font-medium text-sm truncate">{r.title}</div>
                      <div className="text-xs text-brand-600">¥{r.price}</div>
                    </div>
                  </Link>
                ))}
              </div>
            )}
          </div>

          <div className="bg-brand-50 rounded-lg p-4 border border-brand-100">
            <h3 className="font-bold mb-2 text-brand-700">购买须知</h3>
            <ul className="text-xs text-gray-600 leading-relaxed space-y-1 list-disc list-inside">
              <li>团购券请在有效期内使用</li>
              <li>过期未使用可申请退款</li>
              <li>请提前与商家预约</li>
              <li>特价商品不支持退换</li>
            </ul>
          </div>
        </aside>
      </div>
    </div>
  )
}
