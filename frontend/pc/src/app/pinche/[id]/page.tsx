import Link from 'next/link'
import { notFound } from 'next/navigation'
import type { Metadata } from 'next'
import { getPincheDetail, listPinches } from '@/lib/api'

export const revalidate = 0

interface PageProps {
  params: { id: string }
}

export async function generateMetadata({ params }: PageProps): Promise<Metadata> {
  const id = Number(params.id)
  if (Number.isNaN(id)) return { title: '拼车详情' }
  try {
    const item = await getPincheDetail(id)
    return {
      title: `${item.start_location} → ${item.end_location} - 拼车出行`,
      description: item.remark?.slice(0, 120) || '武昌同城拼车出行详情',
      keywords: [item.start_location, item.end_location, '武昌拼车', '顺风车'],
      openGraph: {
        title: `${item.start_location} → ${item.end_location}`,
        description: item.remark?.slice(0, 120),
        type: 'article',
      },
    }
  } catch {
    return { title: '拼车详情' }
  }
}

export default async function PincheDetailPage({ params }: PageProps) {
  const id = Number(params.id)
  if (Number.isNaN(id)) notFound()

  let item
  let related: any[] = []
  try {
    item = await getPincheDetail(id)
  } catch {
    notFound()
  }

  // 侧边栏：相关行程（相同起点/终点）
  try {
    const rel = await listPinches({
      page: 1,
      pageSize: 6,
      startCity: item.start_city,
      endCity: item.end_city,
    })
    related = rel.list.filter((r) => r.id !== item.id).slice(0, 5)
  } catch {
    // 后端不可达
  }

  const isOwner = item.type === 1
  const typeText = isOwner ? '车主找乘客' : '乘客找车主'
  const departure = item.departure_time
    ? new Date(item.departure_time).toLocaleString('zh-CN')
    : ''
  const statusText = item.status === 1 ? '招募中' : item.status === 2 ? '已满员' : '已结束'

  return (
    <div className="container py-6">
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* 主信息 */}
        <div className="lg:col-span-2 bg-white rounded-lg p-6 border border-gray-200">
          <nav className="text-sm text-gray-500 mb-4">
            <Link href="/" className="hover:text-brand-600">首页</Link>
            <span className="mx-2">/</span>
            <Link href="/pinche" className="hover:text-brand-600">拼车出行</Link>
            <span className="mx-2">/</span>
            <span>详情</span>
          </nav>

          <div className="flex items-center justify-between mb-4">
            <h1 className="text-2xl font-bold">
              {item.start_location} → {item.end_location}
            </h1>
            <span className="px-3 py-1 text-sm bg-brand-50 text-brand-700 rounded">{typeText}</span>
          </div>

          {/* 关键信息 */}
          <div className="grid grid-cols-2 sm:grid-cols-4 gap-4 mb-6 py-4 border-y border-gray-100">
            <div>
              <div className="text-xs text-gray-500 mb-1">出发时间</div>
              <div className="font-bold">{departure || '未指定'}</div>
            </div>
            <div>
              <div className="text-xs text-gray-500 mb-1">价格</div>
              <div className="font-bold text-brand-600">¥{item.price}/座</div>
            </div>
            <div>
              <div className="text-xs text-gray-500 mb-1">剩余座位</div>
              <div className="font-bold">{item.seats_left} / {item.seats_total}</div>
            </div>
            <div>
              <div className="text-xs text-gray-500 mb-1">状态</div>
              <div className="font-bold text-brand-600">{statusText}</div>
            </div>
          </div>

          {/* 路线详情 */}
          <div className="mb-6">
            <h3 className="font-bold mb-2">路线详情</h3>
            <div className="space-y-2 text-sm">
              <div className="flex items-center gap-2">
                <span className="w-2 h-2 rounded-full bg-green-500"></span>
                <span className="text-gray-500 w-20">起点：</span>
                <span>{item.start_location}{item.start_city ? `（${item.start_city}）` : ''}</span>
              </div>
              <div className="flex items-center gap-2">
                <span className="w-2 h-2 rounded-full bg-brand-500"></span>
                <span className="text-gray-500 w-20">终点：</span>
                <span>{item.end_location}{item.end_city ? `（${item.end_city}）` : ''}</span>
              </div>
              {item.via && (
                <div className="flex items-center gap-2">
                  <span className="w-2 h-2 rounded-full bg-gray-300"></span>
                  <span className="text-gray-500 w-20">途经：</span>
                  <span>{item.via}</span>
                </div>
              )}
              {item.return_time && (
                <div className="flex items-center gap-2">
                  <span className="w-2 h-2 rounded-full bg-blue-400"></span>
                  <span className="text-gray-500 w-20">返程：</span>
                  <span>{new Date(item.return_time).toLocaleString('zh-CN')}</span>
                </div>
              )}
            </div>
          </div>

          {/* 车辆信息（车主类型） */}
          {isOwner && (item.vehicle_model || item.vehicle_color || item.vehicle_plate) && (
            <div className="mb-6">
              <h3 className="font-bold mb-2">车辆信息</h3>
              <div className="grid grid-cols-1 sm:grid-cols-3 gap-3 text-sm">
                {item.vehicle_model && (
                  <div className="bg-gray-50 p-2 rounded">
                    <div className="text-xs text-gray-500">车型</div>
                    <div className="font-medium">{item.vehicle_model}</div>
                  </div>
                )}
                {item.vehicle_color && (
                  <div className="bg-gray-50 p-2 rounded">
                    <div className="text-xs text-gray-500">颜色</div>
                    <div className="font-medium">{item.vehicle_color}</div>
                  </div>
                )}
                {item.vehicle_plate && (
                  <div className="bg-gray-50 p-2 rounded">
                    <div className="text-xs text-gray-500">车牌</div>
                    <div className="font-medium">{item.vehicle_plate}</div>
                  </div>
                )}
              </div>
            </div>
          )}

          {/* 备注 */}
          {item.remark && (
            <div className="mb-6">
              <h3 className="font-bold mb-2">备注</h3>
              <p className="text-gray-700 whitespace-pre-line">{item.remark}</p>
            </div>
          )}

          {/* 车主/乘客信息 */}
          <div className="flex items-center justify-between pt-4 border-t border-gray-100">
            <div className="flex items-center gap-3">
              <div className="w-10 h-10 rounded-full bg-gray-100 overflow-hidden">
                {item.user_avatar ? (
                  // eslint-disable-next-line @next/next/no-img-element
                  <img src={item.user_avatar} alt={item.user_name || ''} className="w-full h-full object-cover" />
                ) : null}
              </div>
              <div>
                <div className="font-medium text-sm">{item.user_name || '匿名用户'}</div>
                <div className="text-xs text-gray-500">
                  {item.region_name || ''}
                </div>
              </div>
            </div>
            <div className="text-xs text-gray-500">
              👁 {item.view_count || 0} · 📞 {item.contact_count || 0}
            </div>
          </div>
        </div>

        {/* 侧边栏 */}
        <aside className="space-y-4">
          <div className="bg-white rounded-lg p-4 border border-gray-200">
            <h3 className="font-bold mb-3">相关行程</h3>
            {related.length === 0 ? (
              <p className="text-sm text-gray-500">暂无相关行程</p>
            ) : (
              <div className="space-y-3">
                {related.map((r) => (
                  <Link
                    key={r.id}
                    href={`/pinche/${r.id}`}
                    className="block p-2 rounded hover:bg-gray-50"
                  >
                    <div className="font-medium text-sm truncate">
                      {r.start_location} → {r.end_location}
                    </div>
                    <div className="text-xs text-gray-500 flex items-center justify-between mt-1">
                      <span>{r.departure_time ? new Date(r.departure_time).toLocaleDateString('zh-CN') : ''}</span>
                      <span className="text-brand-600">¥{r.price}/座</span>
                    </div>
                  </Link>
                ))}
              </div>
            )}
          </div>

          <div className="bg-brand-50 rounded-lg p-4 border border-brand-100">
            <h3 className="font-bold mb-2 text-brand-700">出行提示</h3>
            <ul className="text-xs text-gray-600 leading-relaxed space-y-1 list-disc list-inside">
              <li>请确认车主身份及车辆信息后再预订</li>
              <li>建议购买出行保险</li>
              <li>遇到纠纷请及时联系平台客服</li>
            </ul>
          </div>
        </aside>
      </div>
    </div>
  )
}
