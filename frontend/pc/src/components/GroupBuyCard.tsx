import Link from 'next/link'
import type { GroupBuy } from '@/lib/types'

export default function GroupBuyCard({ groupbuy }: { groupbuy: GroupBuy }) {
  const discount =
    groupbuy.original_price && groupbuy.original_price > 0
      ? Math.round((groupbuy.price / groupbuy.original_price) * 10)
      : 0

  return (
    <Link
      href={`/groupbuy/${groupbuy.id}`}
      className="block bg-white rounded-lg overflow-hidden border border-gray-200 hover:shadow-md transition-shadow"
    >
      <div className="relative">
        {groupbuy.cover_image ? (
          // eslint-disable-next-line @next/next/no-img-element
          <img
            src={groupbuy.cover_image}
            alt={groupbuy.title}
            className="w-full h-48 object-cover"
          />
        ) : (
          <div className="w-full h-48 bg-gradient-to-br from-orange-50 to-brand-100 flex items-center justify-center text-brand-400 text-sm">
            暂无图片
          </div>
        )}
        {discount > 0 && (
          <span className="absolute top-2 right-2 px-2 py-0.5 text-xs bg-brand-600 text-white rounded">
            {discount}折
          </span>
        )}
      </div>
      <div className="p-3">
        <h3 className="font-bold text-sm mb-1 line-clamp-2 hover:text-brand-600 min-h-[2.5rem]">
          {groupbuy.title}
        </h3>
        {groupbuy.subtitle && (
          <p className="text-xs text-gray-500 line-clamp-1 mb-2">{groupbuy.subtitle}</p>
        )}
        <div className="flex items-center justify-between mb-2">
          <div className="flex items-baseline gap-1">
            <span className="text-brand-600 font-bold text-lg">¥{groupbuy.price}</span>
            {groupbuy.original_price && groupbuy.original_price > groupbuy.price && (
              <span className="text-xs text-gray-400 line-through">
                ¥{groupbuy.original_price}
              </span>
            )}
          </div>
        </div>
        <div className="flex items-center justify-between text-xs text-gray-500">
          <span className="truncate">{groupbuy.shop_name || '商家'}</span>
          <span>已售 {groupbuy.sales_count || 0}</span>
        </div>
      </div>
    </Link>
  )
}
