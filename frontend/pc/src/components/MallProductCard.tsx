import Link from 'next/link'
import type { MallProduct } from '@/lib/types'

export default function MallProductCard({ product }: { product: MallProduct }) {
  return (
    <Link
      href={`/mall/product/${product.id}`}
      className="block bg-white rounded-lg overflow-hidden border border-gray-200 hover:shadow-md transition-shadow"
    >
      {product.cover_image ? (
        // eslint-disable-next-line @next/next/no-img-element
        <img
          src={product.cover_image}
          alt={product.title}
          className="w-full h-48 object-cover"
        />
      ) : (
        <div className="w-full h-48 bg-gray-100 flex items-center justify-center text-gray-400 text-sm">
          暂无图片
        </div>
      )}
      <div className="p-3">
        <h3 className="font-bold text-sm mb-1 line-clamp-2 hover:text-brand-600 min-h-[2.5rem]">
          {product.title}
        </h3>
        {product.subtitle && (
          <p className="text-xs text-gray-500 line-clamp-1 mb-2">{product.subtitle}</p>
        )}
        <div className="flex items-center justify-between mb-2">
          <div className="flex items-baseline gap-1">
            <span className="text-brand-600 font-bold text-lg">¥{product.price}</span>
            {product.original_price && product.original_price > product.price && (
              <span className="text-xs text-gray-400 line-through">
                ¥{product.original_price}
              </span>
            )}
          </div>
        </div>
        <div className="flex items-center justify-between text-xs text-gray-500">
          <span className="truncate">{product.shop_name || '官方店铺'}</span>
          <span>销量 {product.sales_count || 0}</span>
        </div>
      </div>
    </Link>
  )
}
