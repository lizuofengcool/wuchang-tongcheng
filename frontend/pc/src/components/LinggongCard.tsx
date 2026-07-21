import Link from 'next/link'
import type { Linggong } from '@/lib/types'

const TYPE_TEXT: Record<number, string> = { 1: '零工', 2: '兼职', 3: '全职' }
const TYPE_COLOR: Record<number, string> = {
  1: 'bg-blue-100 text-blue-700',
  2: 'bg-purple-100 text-purple-700',
  3: 'bg-brand-100 text-brand-700',
}

export default function LinggongCard({ linggong }: { linggong: Linggong }) {
  const typeText = TYPE_TEXT[linggong.type] || '零工'
  const typeColor = TYPE_COLOR[linggong.type] || 'bg-gray-100 text-gray-700'

  const date = linggong.created_at
    ? new Date(linggong.created_at).toLocaleDateString('zh-CN')
    : ''

  return (
    <Link
      href={`/linggong/${linggong.id}`}
      className="block bg-white rounded-lg p-4 border border-gray-200 hover:shadow-md transition-shadow"
    >
      <div className="flex items-start justify-between gap-3 mb-2">
        <h3 className="font-bold text-base flex-1 line-clamp-1 hover:text-brand-600">
          {linggong.title}
        </h3>
        <span className={`px-2 py-0.5 text-xs rounded whitespace-nowrap ${typeColor}`}>
          {typeText}
        </span>
      </div>

      <div className="text-brand-600 font-bold text-lg mb-2">
        {linggong.salary}
      </div>

      <div className="flex flex-wrap gap-2 text-xs text-gray-500 mb-2">
        {linggong.region_name && <span>📍 {linggong.region_name}</span>}
        {linggong.category_name && <span>🏷 {linggong.category_name}</span>}
        {linggong.work_time && <span>🕒 {linggong.work_time}</span>}
      </div>

      {linggong.description && (
        <p className="text-sm text-gray-600 line-clamp-2 mb-2">{linggong.description}</p>
      )}

      <div className="flex items-center justify-between text-xs text-gray-500 pt-2 border-t border-gray-100">
        <span>{linggong.employer_name || '雇主未公开'}</span>
        <div className="flex items-center gap-3">
          <span>报名 {linggong.applied_count}/{linggong.headcount}</span>
          <span>{date}</span>
        </div>
      </div>
    </Link>
  )
}
