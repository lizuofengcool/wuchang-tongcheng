import Link from 'next/link'
import type { News } from '@/lib/types'

export default function NewsCard({ news }: { news: News }) {
  const date = news.published_at
    ? new Date(news.published_at).toLocaleDateString('zh-CN')
    : ''

  return (
    <Link
      href={`/news/${news.id}`}
      className="block bg-white rounded-lg overflow-hidden border border-gray-200 hover:shadow-md transition-shadow"
    >
      {news.cover_image && (
        // eslint-disable-next-line @next/next/no-img-element
        <img
          src={news.cover_image}
          alt={news.title}
          className="w-full h-48 object-cover"
        />
      )}
      <div className="p-4">
        <h3 className="font-bold text-lg mb-2 line-clamp-2 hover:text-brand-600">
          {news.title}
        </h3>
        {news.summary && (
          <p className="text-sm text-gray-600 line-clamp-2 mb-3">{news.summary}</p>
        )}
        <div className="flex items-center justify-between text-xs text-gray-500">
          <div className="flex items-center gap-3">
            <span>{news.author_name}</span>
            <span>{date}</span>
          </div>
          <div className="flex items-center gap-3">
            <span>👁 {news.view_count}</span>
            <span>❤ {news.like_count}</span>
          </div>
        </div>
      </div>
    </Link>
  )
}
