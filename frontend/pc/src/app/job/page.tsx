import Link from 'next/link'
import type { Metadata } from 'next'
import { listJobs } from '@/lib/api'

export const revalidate = 60

export const metadata: Metadata = {
  title: '招聘求职',
  description:
    '五常同城招聘求职频道，提供本地招聘信息、求职简历、兼职零工等信息。',
  keywords: ['五常招聘', '五常求职', '五常兼职', '五常同城招聘'],
  openGraph: {
    title: '五常同城招聘求职',
    description: '本地招聘信息、求职简历、兼职零工',
    type: 'website',
  },
}

interface SearchParams {
  page?: string
  keyword?: string
}

export default async function JobListPage({
  searchParams,
}: {
  searchParams: SearchParams
}) {
  const page = Number(searchParams.page) || 1
  const keyword = searchParams.keyword || ''

  let jobPage = { list: [] as any[], total: 0, page, pageSize: 12 }
  try {
    jobPage = await listJobs({ page, pageSize: 12, keyword })
  } catch {
    // 后端 /job 接口未就绪，渲染空状态
  }

  const totalPages = Math.ceil(jobPage.total / jobPage.pageSize) || 1

  return (
    <div className="container py-6">
      <nav className="text-sm text-gray-500 mb-4">
        <Link href="/" className="hover:text-brand-600">首页</Link>
        <span className="mx-2">/</span>
        <span>招聘求职</span>
      </nav>

      <h1 className="text-2xl font-bold mb-4">
        招聘求职{keyword ? ` · 搜索：${keyword}` : ''}
      </h1>

      {jobPage.list.length === 0 ? (
        <div className="bg-white rounded-lg p-12 text-center text-gray-500 border border-gray-200">
          暂无招聘信息
          <div className="mt-2 text-xs text-gray-400">
            （后端 /job 接口未就绪时也会展示此状态）
          </div>
        </div>
      ) : (
        <div className="bg-white rounded-lg border border-gray-200 overflow-hidden">
          {jobPage.list.map((item, i) => (
            <Link
              key={item.id}
              href={`/job/${item.id}`}
              className={`block px-4 py-4 hover:bg-gray-50 transition-colors ${
                i > 0 ? 'border-t border-gray-100' : ''
              }`}
            >
              <div className="flex items-center justify-between gap-4">
                <div className="flex-1 min-w-0">
                  <h3 className="font-bold text-base mb-1 truncate hover:text-brand-600">
                    {item.title}
                  </h3>
                  <div className="flex items-center gap-3 text-sm text-gray-500">
                    <span>{item.company}</span>
                    {item.location && <span>📍 {item.location}</span>}
                    {item.category && <span>{item.category}</span>}
                  </div>
                </div>
                <div className="text-brand-600 font-bold whitespace-nowrap">
                  {item.salary}
                </div>
              </div>
            </Link>
          ))}
        </div>
      )}

      {totalPages > 1 && (
        <div className="flex items-center justify-center gap-2 mt-8">
          {page > 1 && (
            <Link
              href={`/job?page=${page - 1}${keyword ? `&keyword=${encodeURIComponent(keyword)}` : ''}`}
              className="px-3 py-1.5 text-sm border rounded hover:bg-gray-50"
            >
              上一页
            </Link>
          )}
          <span className="text-sm text-gray-600">
            第 {page} / {totalPages} 页（共 {jobPage.total} 条）
          </span>
          {page < totalPages && (
            <Link
              href={`/job?page=${page + 1}${keyword ? `&keyword=${encodeURIComponent(keyword)}` : ''}`}
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
