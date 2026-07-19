import Link from 'next/link'
import { notFound } from 'next/navigation'
import type { Metadata } from 'next'
import { getErshou } from '@/lib/api'

export const revalidate = 0

interface PageProps {
  params: { id: string }
}

export async function generateMetadata({ params }: PageProps): Promise<Metadata> {
  const id = Number(params.id)
  if (Number.isNaN(id)) return { title: '二手详情' }
  try {
    const item = await getErshou(id)
    return {
      title: item.title,
      description: item.description?.slice(0, 120) || '五常同城二手物品详情',
      keywords: [item.title, '五常二手', '二手交易'],
      openGraph: {
        title: item.title,
        description: item.description?.slice(0, 120),
        type: 'article',
        images: item.cover_image ? [{ url: item.cover_image }] : undefined,
      },
    }
  } catch {
    return { title: '二手详情' }
  }
}

export default async function ErshouDetailPage({ params }: PageProps) {
  const id = Number(params.id)
  if (Number.isNaN(id)) notFound()

  let item
  try {
    item = await getErshou(id)
  } catch {
    notFound()
  }

  const date = item.created_at
    ? new Date(item.created_at).toLocaleString('zh-CN')
    : ''

  return (
    <div className="container py-6">
      <div className="bg-white rounded-lg p-8 border border-gray-200">
        {/* 面包屑 */}
        <nav className="text-sm text-gray-500 mb-4">
          <Link href="/" className="hover:text-brand-600">首页</Link>
          <span className="mx-2">/</span>
          <Link href="/ershou" className="hover:text-brand-600">二手交易</Link>
          <span className="mx-2">/</span>
          <span>详情</span>
        </nav>

        {/* 标题 */}
        <h1 className="text-3xl font-bold mb-4">{item.title}</h1>

        {/* 元信息 */}
        <div className="flex items-center gap-4 text-sm text-gray-500 mb-6 pb-4 border-b border-gray-100">
          <span className="text-brand-600 font-bold text-xl">¥{item.price}</span>
          {item.original_price && (
            <span className="text-gray-400 line-through">¥{item.original_price}</span>
          )}
          <span>发布者：{item.user_name || '匿名'}</span>
          <span>发布时间：{date}</span>
          <span>👁 {item.view_count || 0}</span>
        </div>

        {/* 封面图 */}
        {item.cover_image && (
          // eslint-disable-next-line @next/next/no-img-element
          <img
            src={item.cover_image}
            alt={item.title}
            className="w-full max-h-96 object-cover rounded mb-6"
          />
        )}

        {/* 多图（如有） */}
        {item.images && item.images.length > 0 && (
          <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 gap-2 mb-6">
            {item.images.map((img, i) => (
              // eslint-disable-next-line @next/next/no-img-element
              <img
                key={i}
                src={img}
                alt={`${item.title} ${i + 1}`}
                className="w-full h-32 object-cover rounded"
              />
            ))}
          </div>
        )}

        {/* 描述 */}
        <div className="news-content">
          <h3 className="font-bold mb-2">物品描述</h3>
          <p>{item.description || '暂无描述'}</p>
        </div>

        {/* 位置 */}
        {item.location && (
          <div className="mt-6 pt-4 border-t border-gray-100">
            <span className="text-sm text-gray-500">📍 {item.location}</span>
          </div>
        )}
      </div>
    </div>
  )
}
