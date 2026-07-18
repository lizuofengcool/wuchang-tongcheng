// 同城头条模块 API 封装
import request from '@/utils/request'

// 头条列表（分页 + 分类/状态/关键词筛选）
export function listNews(params) {
  return request.get('/news', { params })
}

// 头条详情
export function getNews(id) {
  return request.get(`/news/${id}`)
}

// 发布头条
export function createNews(data) {
  return request.post('/news', data)
}

// 更新头条
export function updateNews(id, data) {
  return request.put(`/news/${id}`, data)
}

// 删除头条
export function deleteNews(id) {
  return request.delete(`/news/${id}`)
}

// 点赞/取消点赞（toggle）
export function likeNews(id) {
  return request.post(`/news/${id}/like`)
}

// 查询当前用户对该头条的点赞状态
export function getNewsLikeStatus(id) {
  return request.get(`/news/${id}/like`)
}

// ====== 收藏 ======

// 我的收藏列表（分页，按收藏时间倒序）
export function listFavorites(params) {
  return request.get('/news/favorites', { params })
}

// 收藏/取消收藏（toggle）
export function favNews(id) {
  return request.post(`/news/${id}/fav`)
}

// 查询当前用户对该头条的收藏状态
export function getNewsFavStatus(id) {
  return request.get(`/news/${id}/fav`)
}

// 发布/取消发布新闻
export function toggleNewsStatus(id, status) {
  return request.put(`/news/${id}/status`, { status })
}

// 搜索新闻（ES搜索）
export function searchNews(params) {
  return request.get('/news/search', { params })
}

// ====== 评论 ======

// 评论列表（公开接口，无需登录）
export function listComments(newsId, params) {
  return request.get(`/news/${newsId}/comments`, { params })
}

// 发表评论（需登录）
export function createComment(newsId, data) {
  return request.post(`/news/${newsId}/comments`, data)
}

// 删除评论（需登录，仅作者本人或管理员）
export function deleteComment(commentId) {
  return request.delete(`/news/comments/${commentId}`)
}

// ====== 消息通知 ======

// 我的消息列表（需登录）
export function listMessages(params) {
  return request.get('/news/messages', { params })
}

// 未读消息数（需登录）
export function getUnreadCount() {
  return request.get('/news/messages/unread')
}

// 标记消息已读（需登录，ids 为空数组表示全部已读）
export function markMessagesRead(ids) {
  return request.put('/news/messages/read', { ids: ids || [] })
}
