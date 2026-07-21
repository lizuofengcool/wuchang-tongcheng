import Link from 'next/link'
import { notFound } from 'next/navigation'
import type { Metadata } from 'next'
import { getLoveDetail, listLoves } from '@/lib/api'

export const revalidate = 0

interface PageProps {
  params: { id: string }
}

export async function generateMetadata({ params }: PageProps): Promise<Metadata> {
  const id = Number(params.id)
  if (Number.isNaN(id)) return { title: '相亲详情' }
  try {
    const item = await getLoveDetail(id)
    return {
      title: `${item.nickname || '相亲用户'} - 相亲交友`,
      description: item.bio?.slice(0, 120) || '武昌同城相亲交友详情',
      keywords: [item.nickname || '相亲', '武昌相亲', '同城交友'],
      openGraph: {
        title: `${item.nickname || '相亲用户'} - 相亲交友`,
        description: item.bio?.slice(0, 120),
        type: 'profile',
        images: item.avatar ? [{ url: item.avatar }] : undefined,
      },
    }
  } catch {
    return { title: '相亲详情' }
  }
}

export default async function LoveDetailPage({ params }: PageProps) {
  const id = Number(params.id)
  if (Number.isNaN(id)) notFound()

  let item
  let related: any[] = []
  try {
    item = await getLoveDetail(id)
  } catch {
    notFound()
  }

  // 侧边栏：相关推荐（同性别的其他用户）
  try {
    const relPage = await listLoves({ page: 1, pageSize: 6, gender: item.gender })
    related = relPage.list.filter((r) => r.id !== item.id).slice(0, 5)
  } catch {
    // 后端不可达，省略相关推荐
  }

  const genderText = item.gender === 1 ? '男' : item.gender === 2 ? '女' : '保密'
  const genderColor = item.gender === 1 ? 'bg-blue-100 text-blue-700' : 'bg-pink-100 text-pink-700'

  return (
    <div className="container py-6">
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* 主信息 */}
        <div className="lg:col-span-2 bg-white rounded-lg p-6 border border-gray-200">
          <nav className="text-sm text-gray-500 mb-4">
            <Link href="/" className="hover:text-brand-600">首页</Link>
            <span className="mx-2">/</span>
            <Link href="/love" className="hover:text-brand-600">相亲交友</Link>
            <span className="mx-2">/</span>
            <span>{item.nickname || `用户${item.user_id}`}</span>
          </nav>

          {/* 头部资料 */}
          <div className="flex items-start gap-4 mb-6">
            <div className="w-32 h-32 rounded overflow-hidden bg-gray-100 flex-shrink-0">
              {item.avatar ? (
                // eslint-disable-next-line @next/next/no-img-element
                <img src={item.avatar} alt={item.nickname || '相亲用户'} className="w-full h-full object-cover" />
              ) : (
                <div className="w-full h-full flex items-center justify-center text-gray-400 text-sm">暂无照片</div>
              )}
            </div>
            <div className="flex-1">
              <div className="flex items-center gap-2 mb-2">
                <h1 className="text-2xl font-bold">{item.nickname || `用户${item.user_id}`}</h1>
                <span className={`px-2 py-0.5 text-xs rounded ${genderColor}`}>{genderText}</span>
                {item.verified && (
                  <span className="px-2 py-0.5 text-xs bg-brand-600 text-white rounded">已认证</span>
                )}
              </div>
              <div className="grid grid-cols-2 sm:grid-cols-3 gap-y-2 gap-x-4 text-sm text-gray-600">
                {item.age && <div>年龄：<span className="text-gray-900">{item.age}岁</span></div>}
                {item.height && <div>身高：<span className="text-gray-900">{item.height}cm</span></div>}
                {item.region_name && <div>地区：<span className="text-gray-900">{item.region_name}</span></div>}
                {item.occupation && <div>职业：<span className="text-gray-900">{item.occupation}</span></div>}
                {item.education && <div>学历：<span className="text-gray-900">{item.education}</span></div>}
                {item.income && <div>收入：<span className="text-gray-900">{item.income}</span></div>}
                {item.marriage && <div>婚姻：<span className="text-gray-900">{item.marriage}</span></div>}
                {item.house && <div>住房：<span className="text-gray-900">{item.house}</span></div>}
                {item.car && <div>购车：<span className="text-gray-900">{item.car}</span></div>}
              </div>
            </div>
          </div>

          {/* 个人简介 */}
          {item.bio && (
            <div className="mb-6">
              <h3 className="font-bold mb-2">个人简介</h3>
              <p className="text-gray-700 whitespace-pre-line">{item.bio}</p>
            </div>
          )}

          {/* 标签 */}
          {item.tags && item.tags.length > 0 && (
            <div className="mb-6">
              <h3 className="font-bold mb-2">个人标签</h3>
              <div className="flex flex-wrap gap-2">
                {item.tags.map((tag, i) => (
                  <span key={i} className="px-2 py-1 text-xs bg-gray-100 rounded text-gray-600">
                    #{tag}
                  </span>
                ))}
              </div>
            </div>
          )}

          {/* 照片墙 */}
          {item.photos && item.photos.length > 0 && (
            <div className="mb-6">
              <h3 className="font-bold mb-2">照片墙</h3>
              <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 gap-2">
                {item.photos.map((img, i) => (
                  // eslint-disable-next-line @next/next/no-img-element
                  <img
                    key={i}
                    src={img}
                    alt={`${item.nickname || '相亲用户'} ${i + 1}`}
                    className="w-full h-32 object-cover rounded"
                  />
                ))}
              </div>
            </div>
          )}

          {/* 底部信息 */}
          <div className="flex items-center justify-between text-xs text-gray-500 pt-4 border-t border-gray-100">
            <div className="flex items-center gap-3">
              <span>👁 {item.view_count || 0}</span>
              <span>❤ {item.like_count || 0}</span>
            </div>
            {item.created_at && (
              <span>注册时间：{new Date(item.created_at).toLocaleDateString('zh-CN')}</span>
            )}
          </div>
        </div>

        {/* 侧边栏：相关推荐 */}
        <aside className="space-y-4">
          <div className="bg-white rounded-lg p-4 border border-gray-200">
            <h3 className="font-bold mb-3">相关推荐</h3>
            {related.length === 0 ? (
              <p className="text-sm text-gray-500">暂无推荐用户</p>
            ) : (
              <div className="space-y-3">
                {related.map((r) => (
                  <Link
                    key={r.id}
                    href={`/love/${r.id}`}
                    className="flex items-center gap-3 hover:bg-gray-50 p-2 rounded"
                  >
                    <div className="w-12 h-12 rounded overflow-hidden bg-gray-100 flex-shrink-0">
                      {r.avatar ? (
                        // eslint-disable-next-line @next/next/no-img-element
                        <img src={r.avatar} alt={r.nickname || ''} className="w-full h-full object-cover" />
                      ) : (
                        <div className="w-full h-full flex items-center justify-center text-gray-400 text-xs">无</div>
                      )}
                    </div>
                    <div className="flex-1 min-w-0">
                      <div className="font-medium text-sm truncate">{r.nickname || `用户${r.user_id}`}</div>
                      <div className="text-xs text-gray-500 truncate">
                        {r.age ? `${r.age}岁` : ''} {r.region_name ? `· ${r.region_name}` : ''}
                      </div>
                    </div>
                  </Link>
                ))}
              </div>
            )}
          </div>

          <div className="bg-brand-50 rounded-lg p-4 border border-brand-100">
            <h3 className="font-bold mb-2 text-brand-700">温馨提示</h3>
            <p className="text-xs text-gray-600 leading-relaxed">
              网络交友请注意安全，警惕涉及金钱交易的要求。建议选择实名认证用户，并在公共场合见面。
            </p>
          </div>
        </aside>
      </div>
    </div>
  )
}
