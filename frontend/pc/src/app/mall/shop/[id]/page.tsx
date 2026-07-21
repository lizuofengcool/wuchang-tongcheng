import Link from 'next/link'
import { notFound } from 'next/navigation'
import type { Metadata } from 'next'
import { getMallShopDetail, listMallProducts } from '@/lib/api'
import MallProductCard from '@/components/MallProductCard'

export const revalidate = 0

interface PageProps {
  params: { id: string }
}

export async function generateMetadata({ params }: PageProps): Promise<Metadata> {
  const id = Number(params.id)
  if (Number.isNaN(id)) return { title: '店铺主页' }
  try {
    const shop = await getMallShopDetail(id)
    return {
      title: `${shop.name} - 店铺主页`,
      description: shop.description?.slice(0, 120) || '武昌同城商城店铺主页',
      keywords: [shop.name, '武昌商城', '武昌店铺'],
      openGraph: {
        title: shop.name,
        description: shop.description?.slice(0, 120),
        type: 'article',
        images: shop.logo ? [{ url: shop.logo }] : undefined,
      },
    }
  } catch {
    return { title: '店铺主页' }
  }
}

export default async function MallShopDetailPage({ params }: PageProps) {
  const id = Number(params.id)
  if (Number.isNaN(id)) notFound()

  let shop
  let productPage = { list: [] as any[], total: 0, page: 1, pageSize: 12 }
  try {
    shop = await getMallShopDetail(id)
  } catch {
    notFound()
  }
  try {
    productPage = await listMallProducts({ page: 1, pageSize: 24, shopId: id })
  } catch {
    // 商品接口不可达
  }

  return (
    <div className="container py-6">
      <nav className="text-sm text-gray-500 mb-4">
        <Link href="/" className="hover:text-brand-600">首页</Link>
        <span className="mx-2">/</span>
        <Link href="/mall" className="hover:text-brand-600">同城商城</Link>
        <span className="mx-2">/</span>
        <span>{shop.name}</span>
      </nav>

      {/* 店铺头部 */}
      <div className="bg-white rounded-lg p-6 border border-gray-200 mb-6">
        {shop.banner && (
          // eslint-disable-next-line @next/next/no-img-element
          <img src={shop.banner} alt={shop.name} className="w-full h-40 object-cover rounded mb-4" />
        )}
        <div className="flex items-start gap-4">
          <div className="w-20 h-20 rounded-lg bg-gray-100 overflow-hidden flex-shrink-0">
            {shop.logo ? (
              // eslint-disable-next-line @next/next/no-img-element
              <img src={shop.logo} alt={shop.name} className="w-full h-full object-cover" />
            ) : (
              <div className="w-full h-full flex items-center justify-center text-brand-400 text-2xl font-bold">
                {shop.name?.charAt(0) || '店'}
              </div>
            )}
          </div>
          <div className="flex-1">
            <h1 className="text-2xl font-bold mb-2">{shop.name}</h1>
            {shop.description && (
              <p className="text-sm text-gray-600 mb-2">{shop.description}</p>
            )}
            <div className="flex flex-wrap gap-4 text-xs text-gray-500">
              {shop.category_name && <span>🏷 {shop.category_name}</span>}
              {shop.region_name && <span>📍 {shop.region_name}</span>}
              {shop.address && <span>🏠 {shop.address}</span>}
              {shop.phone && <span>📞 {shop.phone}</span>}
              {shop.rating && <span className="text-brand-600">⭐ {shop.rating.toFixed(1)}</span>}
              <span>📦 {shop.product_count || 0} 件商品</span>
              <span>💬 {shop.review_count || 0} 条评价</span>
            </div>
          </div>
        </div>
      </div>

      {/* 店铺商品 */}
      <div className="mb-4">
        <h2 className="text-xl font-bold mb-4">店铺商品（共 {productPage.total} 件）</h2>
        {productPage.list.length === 0 ? (
          <div className="bg-white rounded-lg p-12 text-center text-gray-500 border border-gray-200">
            暂无商品
          </div>
        ) : (
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">
            {productPage.list.map((item) => (
              <MallProductCard key={item.id} product={item} />
            ))}
          </div>
        )}
      </div>
    </div>
  )
}
