import Link from 'next/link'
import type { Dh114 } from '@/lib/types'

export default function Dh114Card({ dh114 }: { dh114: Dh114 }) {
  const statusText = dh114.status === 1 ? '营业中' : dh114.status === 2 ? '休息中' : '已关闭'
  const statusColor = dh114.status === 1 ? 'bg-green-100 text-green-700' : 'bg-gray-100 text-gray-500'

  return (
    <Link
      href={`/dh114/${dh114.id}`}
      className="block bg-white rounded-lg overflow-hidden border border-gray-200 hover:shadow-md transition-shadow"
    >
      <div className="relative">
        {dh114.cover_image ? (
          // eslint-disable-next-line @next/next/no-img-element
          <img
            src={dh114.cover_image}
            alt={dh114.name}
            className="w-full h-40 object-cover"
          />
        ) : (
          <div className="w-full h-40 bg-gradient-to-br from-brand-50 to-brand-100 flex items-center justify-center text-brand-400 text-3xl font-bold">
            {dh114.name?.charAt(0) || '商'}
          </div>
        )}
        <span className={`absolute top-2 right-2 px-2 py-0.5 text-xs rounded ${statusColor}`}>
          {statusText}
        </span>
      </div>
      <div className="p-3">
        <h3 className="font-bold text-base mb-1 truncate hover:text-brand-600">
          {dh114.name}
        </h3>
        <div className="flex flex-wrap gap-2 text-xs text-gray-500 mb-2">
          {dh114.category_name && <span>🏷 {dh114.category_name}</span>}
          {dh114.region_name && <span>📍 {dh114.region_name}</span>}
        </div>
        {dh114.description && (
          <p className="text-sm text-gray-600 line-clamp-2 mb-2">{dh114.description}</p>
        )}
        <div className="flex items-center justify-between text-xs text-gray-500 pt-2 border-t border-gray-100">
          <div className="flex items-center gap-2">
            {dh114.rating && (
              <span className="text-brand-600">⭐ {dh114.rating.toFixed(1)}</span>
            )}
            {dh114.avg_price && <span>¥{dh114.avg_price}/人</span>}
          </div>
          <span>👁 {dh114.view_count || 0}</span>
        </div>
      </div>
    </Link>
  )
}
