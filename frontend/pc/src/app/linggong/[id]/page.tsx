import Link from 'next/link'
import { notFound } from 'next/navigation'
import type { Metadata } from 'next'
import { getLinggongDetail, listLinggongs, listLinggongTasks } from '@/lib/api'

export const revalidate = 0

interface PageProps {
  params: { id: string }
}

export async function generateMetadata({ params }: PageProps): Promise<Metadata> {
  const id = Number(params.id)
  if (Number.isNaN(id)) return { title: '零工详情' }
  try {
    const item = await getLinggongDetail(id)
    return {
      title: `${item.title} - 零工兼职`,
      description: item.description?.slice(0, 120) || '武昌同城零工兼职详情',
      keywords: [item.title, '武昌零工', '武昌兼职'],
      openGraph: {
        title: item.title,
        description: item.description?.slice(0, 120),
        type: 'article',
      },
    }
  } catch {
    return { title: '零工详情' }
  }
}

export default async function LinggongDetailPage({ params }: PageProps) {
  const id = Number(params.id)
  if (Number.isNaN(id)) notFound()

  let item
  let related: any[] = []
  let tasks: any[] = []
  try {
    item = await getLinggongDetail(id)
  } catch {
    notFound()
  }

  try {
    const rel = await listLinggongs({ page: 1, pageSize: 6, categoryId: item.category_id })
    related = rel.list.filter((r) => r.id !== item.id).slice(0, 5)
  } catch {
    // 后端不可达
  }
  try {
    const t = await listLinggongTasks({ page: 1, pageSize: 10, linggongId: id })
    tasks = t.list || []
  } catch {
    // 任务接口不可达
  }

  const typeText = item.type === 1 ? '零工' : item.type === 2 ? '兼职' : item.type === 3 ? '全职' : '零工'
  const statusText = item.status === 1 ? '招聘中' : item.status === 2 ? '已满员' : '已结束'
  const date = item.created_at ? new Date(item.created_at).toLocaleDateString('zh-CN') : ''

  return (
    <div className="container py-6">
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* 主信息 */}
        <div className="lg:col-span-2 bg-white rounded-lg p-6 border border-gray-200">
          <nav className="text-sm text-gray-500 mb-4">
            <Link href="/" className="hover:text-brand-600">首页</Link>
            <span className="mx-2">/</span>
            <Link href="/linggong" className="hover:text-brand-600">零工兼职</Link>
            <span className="mx-2">/</span>
            <span>详情</span>
          </nav>

          <div className="flex items-center justify-between mb-4">
            <h1 className="text-2xl font-bold">{item.title}</h1>
            <div className="flex items-center gap-2">
              <span className="px-2 py-1 text-xs bg-brand-50 text-brand-700 rounded">{typeText}</span>
              <span className="px-2 py-1 text-xs bg-green-100 text-green-700 rounded">{statusText}</span>
            </div>
          </div>

          {/* 关键信息 */}
          <div className="grid grid-cols-2 sm:grid-cols-4 gap-4 mb-6 py-4 border-y border-gray-100">
            <div>
              <div className="text-xs text-gray-500 mb-1">薪资</div>
              <div className="font-bold text-brand-600 text-lg">{item.salary}</div>
            </div>
            <div>
              <div className="text-xs text-gray-500 mb-1">招聘人数</div>
              <div className="font-bold">{item.headcount} 人</div>
            </div>
            <div>
              <div className="text-xs text-gray-500 mb-1">已报名</div>
              <div className="font-bold">{item.applied_count} 人</div>
            </div>
            <div>
              <div className="text-xs text-gray-500 mb-1">浏览量</div>
              <div className="font-bold">👁 {item.view_count || 0}</div>
            </div>
          </div>

          {/* 工作详情 */}
          <div className="mb-6 grid grid-cols-1 sm:grid-cols-2 gap-y-2 gap-x-6 text-sm">
            {item.region_name && (
              <div><span className="text-gray-500">地区：</span>{item.region_name}</div>
            )}
            {item.address && (
              <div><span className="text-gray-500">地址：</span>{item.address}</div>
            )}
            {item.work_time && (
              <div><span className="text-gray-500">工作时间：</span>{item.work_time}</div>
            )}
            {item.work_duration && (
              <div><span className="text-gray-500">工作周期：</span>{item.work_duration}</div>
            )}
            {item.category_name && (
              <div><span className="text-gray-500">分类：</span>{item.category_name}</div>
            )}
            {item.contact && (
              <div><span className="text-gray-500">联系人：</span>{item.contact}</div>
            )}
            {item.phone && (
              <div><span className="text-gray-500">电话：</span>{item.phone}</div>
            )}
          </div>

          {/* 工作描述 */}
          {item.description && (
            <div className="mb-6">
              <h3 className="font-bold mb-2">工作描述</h3>
              <p className="text-gray-700 whitespace-pre-line">{item.description}</p>
            </div>
          )}

          {/* 要求 */}
          {item.requirements && item.requirements.length > 0 && (
            <div className="mb-6">
              <h3 className="font-bold mb-2">岗位要求</h3>
              <ul className="text-gray-700 space-y-1 list-disc list-inside text-sm">
                {item.requirements.map((req, i) => (
                  <li key={i}>{req}</li>
                ))}
              </ul>
            </div>
          )}

          {/* 福利 */}
          {item.benefits && item.benefits.length > 0 && (
            <div className="mb-6">
              <h3 className="font-bold mb-2">福利待遇</h3>
              <div className="flex flex-wrap gap-2">
                {item.benefits.map((b, i) => (
                  <span key={i} className="px-2 py-1 text-xs bg-green-50 text-green-700 rounded">
                    {b}
                  </span>
                ))}
              </div>
            </div>
          )}

          {/* 任务列表 */}
          {tasks.length > 0 && (
            <div className="mb-6">
              <h3 className="font-bold mb-2">任务进度（{tasks.length}）</h3>
              <div className="space-y-2">
                {tasks.map((t) => (
                  <div key={t.id} className="flex items-center justify-between p-2 bg-gray-50 rounded text-sm">
                    <span className="truncate">{t.title}</span>
                    <span className={`px-2 py-0.5 text-xs rounded ${
                      t.status === 2 ? 'bg-green-100 text-green-700' :
                      t.status === 3 ? 'bg-gray-100 text-gray-500' :
                      'bg-blue-100 text-blue-700'
                    }`}>
                      {t.status === 2 ? '已完成' : t.status === 3 ? '已取消' : '进行中'}
                    </span>
                  </div>
                ))}
              </div>
            </div>
          )}

          {/* 雇主信息 */}
          <div className="flex items-center justify-between pt-4 border-t border-gray-100">
            <div className="flex items-center gap-3">
              <div className="w-10 h-10 rounded-full bg-gray-100 overflow-hidden">
                {item.employer_avatar ? (
                  // eslint-disable-next-line @next/next/no-img-element
                  <img src={item.employer_avatar} alt={item.employer_name || ''} className="w-full h-full object-cover" />
                ) : null}
              </div>
              <div>
                <div className="font-medium text-sm">{item.employer_name || '雇主未公开'}</div>
                <div className="text-xs text-gray-500">发布于 {date}</div>
              </div>
            </div>
          </div>
        </div>

        {/* 侧边栏 */}
        <aside className="space-y-4">
          <div className="bg-white rounded-lg p-4 border border-gray-200">
            <h3 className="font-bold mb-3">相关岗位</h3>
            {related.length === 0 ? (
              <p className="text-sm text-gray-500">暂无相关岗位</p>
            ) : (
              <div className="space-y-3">
                {related.map((r) => (
                  <Link
                    key={r.id}
                    href={`/linggong/${r.id}`}
                    className="block p-2 rounded hover:bg-gray-50"
                  >
                    <div className="font-medium text-sm truncate">{r.title}</div>
                    <div className="text-xs text-gray-500 flex items-center justify-between mt-1">
                      <span>{r.employer_name || '雇主'}</span>
                      <span className="text-brand-600">{r.salary}</span>
                    </div>
                  </Link>
                ))}
              </div>
            )}
          </div>

          <div className="bg-brand-50 rounded-lg p-4 border border-brand-100">
            <h3 className="font-bold mb-2 text-brand-700">报名提示</h3>
            <ul className="text-xs text-gray-600 leading-relaxed space-y-1 list-disc list-inside">
              <li>报名前请确认岗位要求和工作地点</li>
              <li>警惕需要预交费用的岗位</li>
              <li>建议保留沟通记录和合同</li>
            </ul>
          </div>
        </aside>
      </div>
    </div>
  )
}
