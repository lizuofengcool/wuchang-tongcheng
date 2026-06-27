import Link from 'next/link'
import { notFound } from 'next/navigation'
import { getNews } from '@/lib/api'
import LikeButton from './LikeButton'

export const revalidate = 0 // 详情不缓存（每次访问都更新浏览量）

interface PageProps {
  params: { id: string }
}

export default async function NewsDetailPage({ params }: PageProps) {
  const id = Number(params.id)
  if (Number.isNaN(id)) notFound()

  let news
  try {
    news = await getNews(id)
  } catch {
    notFound()
  }

  const date = news.published_at
    ? new Date(news.published_at).toLocaleString('zh-CN')
    : ''

  return (
    <div className="container py-6">
      <div className="bg-white rounded-lg p-8 border border-gray-200">
        {/* 面包屑 */}
        <nav className="text-sm text-gray-500 mb-4">
          <Link href="/" className="hover:text-brand-600">首页</Link>
          <span className="mx-2">/</span>
          <Link href="/news" className="hover:text-brand-600">同城头条</Link>
          <span className="mx-2">/</span>
          <span>详情</span>
        </nav>

        {/* 标题 */}
        <h1 className="text-3xl font-bold mb-4">{news.title}</h1>

        {/* 元信息 */}
        <div className="flex items-center gap-4 text-sm text-gray-500 mb-6 pb-4 border-b border-gray-100">
          <span>作者：{news.author_name}</span>
          <span>发布时间：{date}</span>
          <span>👁 浏览 {news.view_count}</span>
          <span>❤ 点赞 {news.like_count}</span>
        </div>

        {/* 封面图 */}
        {news.cover_image && (
          // eslint-disable-next-line @next/next/no-img-element
          <img
            src={news.cover_image}
            alt={news.title}
            className="w-full max-h-96 object-cover rounded mb-6"
          />
        )}

        {/* 正文（后端是富文本 HTML） */}
        <div
          className="news-content"
          dangerouslySetInnerHTML={{ __html: news.content }}
        />

        {/* 标签 */}
        {news.tags && (
          <div className="mt-8 pt-4 border-t border-gray-100">
            <div className="flex flex-wrap gap-2">
              {news.tags.split(',').filter(Boolean).map((tag, i) => (
                <span key={i} className="px-2 py-1 text-xs bg-gray-100 rounded text-gray-600">
                  #{tag.trim()}
                </span>
              ))}
            </div>
          </div>
        )}

        {/* 点赞按钮（客户端组件，需登录） */}
        <div className="mt-8 pt-4 border-t border-gray-100">
          <LikeButton newsId={news.id} initialLikeCount={news.like_count} />
        </div>
      </div>
    </div>
  )
}
