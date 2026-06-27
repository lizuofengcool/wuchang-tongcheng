'use client'

import { useState } from 'react'

// 点赞按钮（客户端组件）
// PC门户默认未登录：点击后引导用户去管理后台登录（PC门户暂未实现登录页）
// 如已登录（localStorage 存有 token），则调用点赞 API

export default function LikeButton({
  newsId,
  initialLikeCount,
}: {
  newsId: number
  initialLikeCount: number
}) {
  const [liked, setLiked] = useState(false)
  const [likeCount, setLikeCount] = useState(initialLikeCount)
  const [loading, setLoading] = useState(false)
  const [msg, setMsg] = useState('')

  const getToken = (): string | null => {
    if (typeof window === 'undefined') return null
    return localStorage.getItem('token')
  }

  const onClick = async () => {
    const token = getToken()
    if (!token) {
      setMsg('请先登录后点赞（前往管理后台 /login 登录）')
      return
    }
    setLoading(true)
    setMsg('')
    try {
      const res = await fetch(`/api/v1/news/${newsId}/like`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${token}`,
        },
      })
      const json = await res.json()
      if (json.code !== 0) {
        setMsg(json.message || '点赞失败')
        return
      }
      setLiked(json.data.liked)
      setLikeCount(json.data.like_count)
    } catch (e) {
      setMsg('网络错误')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="flex items-center gap-3">
      <button
        onClick={onClick}
        disabled={loading}
        className={`px-6 py-2 rounded-full border transition-all ${
          liked
            ? 'bg-brand-600 text-white border-brand-600'
            : 'bg-white text-brand-600 border-brand-600 hover:bg-brand-50'
        } ${loading ? 'opacity-50' : ''}`}
      >
        {liked ? '❤ 已点赞' : '🤍 点赞'}（{likeCount}）
      </button>
      {msg && <span className="text-sm text-gray-500">{msg}</span>}
    </div>
  )
}
