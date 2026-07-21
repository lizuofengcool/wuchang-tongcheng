import Link from 'next/link'
import { notFound } from 'next/navigation'
import type { Metadata } from 'next'
import { getMallProductDetail, getMallShopDetail, listMallProducts } from '@/lib/api'

export const revalidate = 0

interface PageProps {
  params: { id: string }
}

export async function generateMetadata({ params }: PageProps): Promise<Metadata> {
  const id = Number(params.id)
  if (Number.isNaN(id)) return { title: '商品详情' }
  try {
    const item = await getMallProductDetail(id)
    return {
      title: `${item.title} - 同城商城`,
      description: item.subtitle?.slice(0, 120) || item.description?.slice(0, 120) || '武昌同城商城商品详情',
      keywords: [item.title, '武昌商城', '武昌购物'],
      openGraph: {
        title: item.title,
        description: item.subtitle?.slice(0, 120),
        type: 'article',
        images: item.cover_image ? [{ url: item.cover_image }] : undefined,
      },
    }
  } catch {
    return { title: '商品详情' }
  }
}

export default async function MallProductDetailPage({ params }: PageProps) {
  const id = Number(params.id)
  if (Number.isNaN(id)) notFound()

  let item
  let shop: any = null
  let related: any[] = []
  try {
    item = await getMallProductDetail(id)
  } catch {
    notFound()
  }

  try {
    shop = await getMallShopDetail(item.shop_id)
  } catch {
    // 店铺接口不可达
  }
  try {
    const rel = await listMallProducts({ page: 1, pageSize: 6, categoryId: item.category_id })
    related = rel.list.filter((r) => r.id !== item.id).slice(0, 5)
  } catch {
    // 后端不可达
  }

  return (
    <div className="container py-6">
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* 主信息 */}
        <div className="lg:col-span-2 bg-white rounded-lg p-6 border border-gray-200">
          <nav className="text-sm text-gray-500 mb-4">
            <Link href="/" className="hover:text-brand-600">首页</Link>
            <span className="mx-2">/</span>
            <Link href="/mall" className="hover:text-brand-600">同城商城</Link>
            <span className="mx-2">/</span>
            <span>{item.title}</span>
          </nav>

          {/* 商品图 + 信息 */}
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-6 mb-6">
            <div className="bg-gray-50 rounded-lg overflow-hidden">
              {item.cover_image ? (
                // eslint-disable-next-line @next/next/no-img-element
                <img src={item.cover_image} alt={item.title} className="w-full h-80 object-cover" />
              ) : (
                <div className="w-full h-80 flex items-center justify-center text-gray-400 text-sm">暂无图片</div>
              )}
            </div>
            <div>
              <h1 className="text-xl font-bold mb-2">{item.title}</h1>
              {item.subtitle && (
                <p className="text-sm text-gray-500 mb-3">{item.subtitle}</p>
              )}

              {/* 价格 */}
              <div className="bg-brand-50 p-4 rounded mb-4">
                <div className="flex items-baseline gap-2">
                  <span className="text-brand-600 font-bold text-2xl">¥{item.price}</span>
                  {item.original_price && item.original_price > item.price && (
                    <span className="text-sm text-gray-400 line-through">¥{item.original_price}</span>
                  )}
                </div>
              </div>

              {/* 关键信息 */}
              <div className="grid grid-cols-2 gap-y-2 gap-x-4 text-sm text-gray-600 mb-4">
                <div>销量：<span className="text-gray-900">{item.sales_count || 0}</span></div>
                <div>库存：<span className="text-gray-900">{item.stock}</span></div>
                <div>浏览：<span className="text-gray-900">{item.view_count || 0}</span></div>
                {item.rating && <div>评分：<span className="text-brand-600">⭐ {item.rating.toFixed(1)}</span></div>}
                {item.category_name && <div>分类：<span className="text-gray-900">{item.category_name}</span></div>}
              </div>

              {/* 购买按钮（仅展示，PC门户无下单流程） */}
              <div className="flex items-center gap-3">
                <button
                  type="button"
                  disabled
                  className="px-6 py-2 bg-brand-600 text-white rounded hover:bg-brand-700 disabled:opacity-50"
                >
                  立即购买
                </button>
                <button
                  type="button"
                  disabled
                  className="px-6 py-2 border border-brand-600 text-brand-600 rounded hover:bg-brand-50 disabled:opacity-50"
                >
                  加入购物车
                </button>
                <span className="text-xs text-gray-400">请登录后操作</span>
              </div>
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

          {/* 商品详情 */}
          {item.description && (
            <div className="mb-6">
              <h3 className="font-bold mb-2">商品详情</h3>
              <div className="text-gray-700 whitespace-pre-line text-sm leading-relaxed">
                {item.description}
              </div>
            </div>
          )}

          {/* 店铺信息 */}
          {shop && (
            <div className="mt-6 pt-4 border-t border-gray-100">
              <Link
                href={`/mall/shop/${shop.id}`}
                className="flex items-center gap-3 hover:bg-gray-50 p-3 rounded"
              >
                <div className="w-12 h-12 rounded bg-gray-100 overflow-hidden flex-shrink-0">
                  {shop.logo ? (
                    // eslint-disable-next-line @next/next/no-img-element
                    <img src={shop.logo} alt={shop.name} className="w-full h-full object-cover" />
                  ) : (
                    <div className="w-full h-full flex items-center justify-center text-brand-400 font-bold">
                      {shop.name?.charAt(0) || '店'}
                    </div>
                  )}
                </div>
                <div className="flex-1 min-w-0">
                  <div className="font-medium text-sm">{shop.name}</div>
                  <div className="text-xs text-gray-500">
                    {shop.category_name || ''} · {shop.product_count || 0} 件商品
                    {shop.rating && <span className="ml-2 text-brand-600">⭐ {shop.rating.toFixed(1)}</span>}
                  </div>
                </div>
                <span className="text-sm text-brand-600">进店 →</span>
              </Link>
            </div>
          )}
        </div>

        {/* 侧边栏 */}
        <aside className="space-y-4">
          <div className="bg-white rounded-lg p-4 border border-gray-200">
            <h3 className="font-bold mb-3">相关商品</h3>
            {related.length === 0 ? (
              <p className="text-sm text-gray-500">暂无相关商品</p>
            ) : (
              <div className="space-y-3">
                {related.map((r) => (
                  <Link
                    key={r.id}
                    href={`/mall/product/${r.id}`}
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
        </aside>
      </div>
    </div>
  )
}
