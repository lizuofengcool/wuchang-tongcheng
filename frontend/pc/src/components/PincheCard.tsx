import Link from 'next/link'
import type { Pinche } from '@/lib/types'

export default function PincheCard({ pinche }: { pinche: Pinche }) {
  const isOwner = pinche.type === 1
  const typeText = isOwner ? '车主找乘客' : '乘客找车主'
  const typeColor = isOwner ? 'bg-green-100 text-green-700' : 'bg-orange-100 text-orange-700'

  const departure = pinche.departure_time
    ? new Date(pinche.departure_time).toLocaleString('zh-CN', {
        month: '2-digit',
        day: '2-digit',
        hour: '2-digit',
        minute: '2-digit',
      })
    : ''

  return (
    <Link
      href={`/pinche/${pinche.id}`}
      className="block bg-white rounded-lg p-4 border border-gray-200 hover:shadow-md transition-shadow"
    >
      <div className="flex items-center justify-between mb-2">
        <span className={`px-2 py-0.5 text-xs rounded ${typeColor}`}>{typeText}</span>
        <span className="text-brand-600 font-bold">¥{pinche.price}/座</span>
      </div>

      <div className="flex items-center gap-2 mb-2">
        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2 text-sm">
            <span className="w-2 h-2 rounded-full bg-green-500"></span>
            <span className="truncate font-medium">{pinche.start_location}</span>
          </div>
          <div className="flex items-center gap-2 text-sm mt-1">
            <span className="w-2 h-2 rounded-full bg-brand-500"></span>
            <span className="truncate font-medium">{pinche.end_location}</span>
          </div>
        </div>
      </div>

      <div className="flex items-center justify-between text-xs text-gray-500">
        <span>🕒 {departure}</span>
        <span>座位 {pinche.seats_left}/{pinche.seats_total}</span>
      </div>

      {(pinche.vehicle_model || pinche.remark) && (
        <div className="mt-2 pt-2 border-t border-gray-100 text-xs text-gray-500 line-clamp-1">
          {pinche.vehicle_model && <span>🚗 {pinche.vehicle_model} · </span>}
          {pinche.remark && <span>{pinche.remark}</span>}
        </div>
      )}
    </Link>
  )
}
