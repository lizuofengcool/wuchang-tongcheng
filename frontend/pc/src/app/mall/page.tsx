import Link from 'next/link'
import type { Metadata } from 'next'
import { listMallProducts, listMallShops, listMallCategories } from '@/lib/api'
import MallProductCard from '@/components/MallProductCard'

export const revalidate = 60

export const metadata: Metadata = {
  title: '同城商城',
  description:
    '武昌同城商城频道，提供本地商品、店铺、特产、生活用品在线购买，同城速达，品质保障。',
  keywords: ['武昌商城', '武昌购物', '同城商城', '武昌特产'],
  openGraph: {
    title: '武昌同城商城',
    description: '本地商品、店铺、特产在线购买',
    type: 'website',
  },
}

interface SearchParams {
  page?: string
  categoryId?: string
  keyword?: string
}

export default async function MallHomePage({
  searchParams,
}: {
  searchParams: SearchParams
}) {
  const page = Number(searchParams.page) || 1
  const categoryId = searchParams.categoryId ? Number(searchParams.categoryId) : undefined
  const keyword = searchParams.keyword || ''

  let productPage = { list: [] as any[], total: 0, page, pageSize: 12 }
  let shopPage = { list: [] as any[], total: 0, page: 1, pageSize: 8 }
  let categories: any[] = []
  try {
    productPage = await listMallProducts({ page, pageSize: 12, categoryId, keyword })
  } catch {
    // 后端 /mall/products 接口未就绪
  }
  try {
    shopPage = await listMallShops({ page: 1, pageSize: 8 })
  } catch {
    // 店铺接口不可达
  }
  try {
    categories = await listMallCategories()
  } catch {
    // 分类接口不可达
  }

  const totalPages = Math.ceil(productPage.total / productPage.pageSize) || 1

  const buildPageHref = (p: number) => {
    const q = new URLSearchParams()
    if (p !== 1) q.set('page', String(p))
    if (categoryId) q.set('categoryId', String(categoryId))
    if (keyword) q.set('keyword', keyword)
    const s = q.toString()
    return s ? `/mall?${s}` : '/mall'
  }

  return (
    <div className="container py-6">
      <nav className="text-sm text-gray-500 mb-4">
        <Link href="/" className="hover:text-brand-600">首页</Link>
        <span className="mx-2">/</span>
        <span>同城商城</span>
      </nav>

      <h1 className="text-2xl font-bold mb-4">
        同城商城{keyword ? ` · 搜索：${keyword}` : ''}
      </h1>

      {/* 分类导航 */}
      {categories.length > 0 && (
        <div className="bg-white rounded-lg p-4 mb-6 border border-gray-200">
          <h3 className="font-bold mb-3 text-sm">分类导航</h3>
          <div className="flex flex-wrap gap-2 text-sm">
            <Link
              href="/mall"
              className={`px-3 py-1.5 rounded ${!categoryId ? 'bg-brand-600 text-white' : 'bg-gray-100 hover:bg-brand-50 hover:text-brand-600'}`}
            >
              全部
            </Link>
            {categories.map((c) => (
              <Link
                key={c.id}
                href={`/mall?categoryId=${c.id}`}
                className={`px-3 py-1.5 rounded ${categoryId === c.id ? 'bg-brand-600 text-white' : 'bg-gray-100 hover:bg-brand-50 hover:text-brand-600'}`}
              >
                {c.name}
              </Link>
            ))}
          </div>
        </div>
      )}

      {/* 搜索 */}
      <div className="bg-white rounded-lg p-4 mb-6 border border-gray-200">
        <form className="flex items-center gap-2" method="get" action="/mall">
          <input
            type="text"
            name="keyword"
            defaultValue={keyword}
            placeholder="搜索商品…"
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

      {/* 商品列表 */}
      {productPage.list.length === 0 ? (
        <div className="bg-white rounded-lg p-12 text-center text-gray-500 border border-gray-200">
          暂无商品
          <div className="mt-2 text-xs text-gray-400">
            （后端 /mall/products 接口未就绪时也会展示此状态）
          </div>
        </div>
      ) : (
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">
          {productPage.list.map((item) => (
            <MallProductCard key={item.id} product={item} />
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
            第 {page} / {totalPages} 页（共 {productPage.total} 件商品）
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

      {/* 推荐店铺 */}
      {shopPage.list.length > 0 && (
        <section className="mt-8">
          <div className="flex items-center justify-between mb-4">
            <h2 className="text-xl font-bold">推荐店铺</h2>
          </div>
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
            {shopPage.list.map((shop) => (
              <Link
                key={shop.id}
                href={`/mall/shop/${shop.id}`}
                className="block bg-white rounded-lg p-4 border border-gray-200 hover:shadow-md transition-shadow"
              >
                <div className="flex items-center gap-3 mb-2">
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
                    <div className="font-medium text-sm truncate hover:text-brand-600">{shop.name}</div>
                    <div className="text-xs text-gray-500">
                      {shop.category_name || ''} · {shop.product_count || 0} 件商品
                    </div>
                  </div>
                </div>
                {shop.description && (
                  <p className="text-xs text-gray-500 line-clamp-2">{shop.description}</p>
                )}
              </Link>
            ))}
          </div>
        </section>
      )}
    </div>
  )
}
