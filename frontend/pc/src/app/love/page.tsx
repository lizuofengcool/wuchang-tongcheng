import Link from 'next/link'
import type { Metadata } from 'next'
import { listLoves } from '@/lib/api'
import LoveCard from '@/components/LoveCard'

export const revalidate = 60

export const metadata: Metadata = {
  title: '相亲交友',
  description:
    '武昌同城相亲交友频道，提供本地真实相亲信息，实名认证、智能匹配，让同城有缘人更易相遇。',
  keywords: ['武昌相亲', '同城交友', '武昌婚恋', '相亲交友'],
  openGraph: {
    title: '武昌同城相亲交友',
    description: '本地真实相亲信息，实名认证，智能匹配',
    type: 'website',
  },
}

interface SearchParams {
  page?: string
  gender?: string
  ageRange?: string
  keyword?: string
}

export default async function LoveListPage({
  searchParams,
}: {
  searchParams: SearchParams
}) {
  const page = Number(searchParams.page) || 1
  const gender = searchParams.gender ? Number(searchParams.gender) : undefined
  const ageRange = searchParams.ageRange || ''
  const keyword = searchParams.keyword || ''

  let lovePage = { list: [] as any[], total: 0, page, pageSize: 12 }
  try {
    lovePage = await listLoves({ page, pageSize: 12, gender, ageRange, keyword })
  } catch {
    // 后端 /love 接口未就绪，渲染空状态
  }

  const totalPages = Math.ceil(lovePage.total / lovePage.pageSize) || 1

  const buildPageHref = (p: number) => {
    const q = new URLSearchParams()
    if (p !== 1) q.set('page', String(p))
    if (gender) q.set('gender', String(gender))
    if (ageRange) q.set('ageRange', ageRange)
    if (keyword) q.set('keyword', keyword)
    const s = q.toString()
    return s ? `/love?${s}` : '/love'
  }

  return (
    <div className="container py-6">
      <nav className="text-sm text-gray-500 mb-4">
        <Link href="/" className="hover:text-brand-600">首页</Link>
        <span className="mx-2">/</span>
        <span>相亲交友</span>
      </nav>

      <div className="flex items-center justify-between mb-4">
        <h1 className="text-2xl font-bold">
          相亲交友{keyword ? ` · 搜索：${keyword}` : ''}
        </h1>
      </div>

      {/* 筛选区 */}
      <div className="bg-white rounded-lg p-4 mb-6 border border-gray-200">
        <div className="flex flex-wrap items-center gap-4 text-sm">
          <div className="flex items-center gap-2">
            <span className="text-gray-500">性别：</span>
            <Link
              href={buildPageHref(1).replace(/gender=\d+&?/, '').replace(/\?$/, '') || '/love'}
              className={`px-2 py-1 rounded ${!gender ? 'bg-brand-600 text-white' : 'hover:bg-gray-100'}`}
            >
              不限
            </Link>
            <Link
              href={`${buildPageHref(1)}${buildPageHref(1).includes('?') ? '&' : '?'}gender=1`}
              className={`px-2 py-1 rounded ${gender === 1 ? 'bg-brand-600 text-white' : 'hover:bg-gray-100'}`}
            >
              男
            </Link>
            <Link
              href={`${buildPageHref(1)}${buildPageHref(1).includes('?') ? '&' : '?'}gender=2`}
              className={`px-2 py-1 rounded ${gender === 2 ? 'bg-brand-600 text-white' : 'hover:bg-gray-100'}`}
            >
              女
            </Link>
          </div>

          <div className="flex items-center gap-2">
            <span className="text-gray-500">年龄：</span>
            {['', '18-25', '26-30', '31-35', '36-45', '46+'].map((ar) => (
              <Link
                key={ar || 'all'}
                href={`${buildPageHref(1)}${buildPageHref(1).includes('?') ? '&' : '?'}${ar ? `ageRange=${ar}` : ''}`}
                className={`px-2 py-1 rounded ${ageRange === ar ? 'bg-brand-600 text-white' : 'hover:bg-gray-100'}`}
              >
                {ar || '不限'}
              </Link>
            ))}
          </div>

          <form className="flex items-center gap-2 ml-auto" method="get" action="/love">
            <input
              type="text"
              name="keyword"
              defaultValue={keyword}
              placeholder="搜索昵称/职业/简介…"
              className="w-48 px-3 py-1.5 text-sm border border-gray-300 rounded focus:outline-none focus:border-brand-500"
            />
            <button
              type="submit"
              className="px-3 py-1.5 text-sm bg-brand-600 text-white rounded hover:bg-brand-700"
            >
              搜索
            </button>
          </form>
        </div>
      </div>

      {lovePage.list.length === 0 ? (
        <div className="bg-white rounded-lg p-12 text-center text-gray-500 border border-gray-200">
          暂无相亲用户
          <div className="mt-2 text-xs text-gray-400">
            （后端 /love 接口未就绪时也会展示此状态）
          </div>
        </div>
      ) : (
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">
          {lovePage.list.map((item) => (
            <LoveCard key={item.id} love={item} />
          ))}
        </div>
      )}

      {totalPages > 1 && (
        <div className="flex items-center justify-center gap-2 mt-8">
          {page > 1 && (
            <Link
              href={buildPageHref(page - 1)}
              className="px-3 py-1.5 text-sm border rounded hover:bg-gray-50"
            >
              上一页
            </Link>
          )}
          <span className="text-sm text-gray-600">
            第 {page} / {totalPages} 页（共 {lovePage.total} 位用户）
          </span>
          {page < totalPages && (
            <Link
              href={buildPageHref(page + 1)}
              className="px-3 py-1.5 text-sm border rounded hover:bg-gray-50"
            >
              下一页
            </Link>
          )}
        </div>
      )}
    </div>
  )
}
