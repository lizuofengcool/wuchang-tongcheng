'use client'

import Link from 'next/link'
import { useState } from 'react'
import { useRouter } from 'next/navigation'

export default function Header() {
  const router = useRouter()
  const [keyword, setKeyword] = useState('')

  const onSearch = (e: React.FormEvent) => {
    e.preventDefault()
    if (keyword.trim()) {
      router.push(`/news/search?keyword=${encodeURIComponent(keyword.trim())}`)
    }
  }

  return (
    <header className="bg-white border-b border-gray-200 sticky top-0 z-10">
      <div className="container flex items-center h-16">
        <Link href="/" className="flex items-center gap-2 mr-8">
          <span className="text-2xl font-bold text-brand-600">五常同城</span>
          <span className="text-xs text-gray-500 hidden sm:inline">本地生活服务平台</span>
        </Link>

        <nav className="flex-1 flex items-center gap-6 text-sm">
          <Link href="/" className="hover:text-brand-600">首页</Link>
          <Link href="/news" className="hover:text-brand-600">同城头条</Link>
          <Link href="/news/search" className="hover:text-brand-600">搜索</Link>
        </nav>

        <form onSubmit={onSearch} className="flex items-center gap-2">
          <input
            type="text"
            value={keyword}
            onChange={(e) => setKeyword(e.target.value)}
            placeholder="搜索头条、商家、服务…"
            className="w-48 sm:w-64 px-3 py-1.5 text-sm border border-gray-300 rounded focus:outline-none focus:border-brand-500"
          />
          <button
            type="submit"
            className="px-4 py-1.5 text-sm bg-brand-600 text-white rounded hover:bg-brand-700"
          >
            搜索
          </button>
        </form>
      </div>
    </header>
  )
}
