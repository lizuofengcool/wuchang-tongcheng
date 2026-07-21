import Link from 'next/link'
import type { Love } from '@/lib/types'

const GENDER_TEXT: Record<number, string> = { 1: '男', 2: '女' }

export default function LoveCard({ love }: { love: Love }) {
  const genderText = GENDER_TEXT[love.gender] || '保密'
  const genderColor = love.gender === 1 ? 'bg-blue-100 text-blue-700' : 'bg-pink-100 text-pink-700'

  return (
    <Link
      href={`/love/${love.id}`}
      className="block bg-white rounded-lg overflow-hidden border border-gray-200 hover:shadow-md transition-shadow"
    >
      <div className="relative">
        {love.avatar ? (
          // eslint-disable-next-line @next/next/no-img-element
          <img
            src={love.avatar}
            alt={love.nickname || '相亲用户'}
            className="w-full h-64 object-cover"
          />
        ) : (
          <div className="w-full h-64 bg-gray-100 flex items-center justify-center text-gray-400 text-sm">
            暂无照片
          </div>
        )}
        <span className={`absolute top-2 right-2 px-2 py-0.5 text-xs rounded ${genderColor}`}>
          {genderText}
        </span>
        {love.verified && (
          <span className="absolute top-2 left-2 px-2 py-0.5 text-xs bg-brand-600 text-white rounded">
            已认证
          </span>
        )}
      </div>
      <div className="p-3">
        <div className="flex items-center justify-between mb-1">
          <h3 className="font-bold text-base truncate hover:text-brand-600">
            {love.nickname || `用户${love.user_id}`}
          </h3>
          <span className="text-sm text-gray-500">{love.age ? `${love.age}岁` : ''}</span>
        </div>
        <div className="flex flex-wrap gap-1 text-xs text-gray-500 mb-2">
          {love.region_name && <span>📍 {love.region_name}</span>}
          {love.occupation && <span>💼 {love.occupation}</span>}
          {love.height && <span>📏 {love.height}cm</span>}
        </div>
        {love.bio && (
          <p className="text-sm text-gray-600 line-clamp-2">{love.bio}</p>
        )}
        <div className="flex items-center justify-between text-xs text-gray-500 mt-2">
          <span>👁 {love.view_count || 0}</span>
          <span>❤ {love.like_count || 0}</span>
        </div>
      </div>
    </Link>
  )
}
