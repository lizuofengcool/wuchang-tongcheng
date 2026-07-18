// 单元测试：src/api/news.js
// 验证每个 API 函数调用 request 的方法、URL、入参
import { describe, it, expect, beforeEach, vi } from 'vitest'

const requestMock = vi.fn()
vi.mock('@/utils/request', () => ({
  default: {
    get: (...a) => requestMock('get', ...a),
    post: (...a) => requestMock('post', ...a),
    put: (...a) => requestMock('put', ...a),
    delete: (...a) => requestMock('delete', ...a)
  }
}))

import * as newsApi from '../news'

beforeEach(() => {
  requestMock.mockReset()
  requestMock.mockImplementation(() => Promise.resolve({ code: 0, data: {} }))
})

describe('news API - 列表/详情/CRUD', () => {
  it('listNews → GET /news，params 透传', async () => {
    const params = { page: 1, page_size: 10, category_id: 3, status: 1, keyword: 'foo' }
    await newsApi.listNews(params)
    expect(requestMock).toHaveBeenCalledWith('get', '/news', { params })
  })

  it('listNews 无参 → params undefined 仍可调用', async () => {
    await newsApi.listNews()
    expect(requestMock).toHaveBeenCalledWith('get', '/news', { params: undefined })
  })

  it('getNews → GET /news/:id', async () => {
    await newsApi.getNews(88)
    expect(requestMock).toHaveBeenCalledWith('get', '/news/88')
  })

  it('createNews → POST /news，body 透传', async () => {
    const data = { title: 't', content: 'c', category_id: 1 }
    await newsApi.createNews(data)
    expect(requestMock).toHaveBeenCalledWith('post', '/news', data)
  })

  it('updateNews → PUT /news/:id，body 透传', async () => {
    const data = { title: 't2' }
    await newsApi.updateNews(88, data)
    expect(requestMock).toHaveBeenCalledWith('put', '/news/88', data)
  })

  it('deleteNews → DELETE /news/:id', async () => {
    await newsApi.deleteNews(88)
    expect(requestMock).toHaveBeenCalledWith('delete', '/news/88')
  })

  it('toggleNewsStatus → PUT /news/:id/status，body 仅含 status', async () => {
    await newsApi.toggleNewsStatus(88, 1)
    expect(requestMock).toHaveBeenCalledWith('put', '/news/88/status', { status: 1 })
  })

  it('searchNews → GET /news/search，params 透传', async () => {
    const params = { keyword: 'foo', page: 1, page_size: 20 }
    await newsApi.searchNews(params)
    expect(requestMock).toHaveBeenCalledWith('get', '/news/search', { params })
  })
})

describe('news API - 点赞', () => {
  it('likeNews → POST /news/:id/like（toggle）', async () => {
    await newsApi.likeNews(88)
    expect(requestMock).toHaveBeenCalledWith('post', '/news/88/like')
  })

  it('getNewsLikeStatus → GET /news/:id/like', async () => {
    await newsApi.getNewsLikeStatus(88)
    expect(requestMock).toHaveBeenCalledWith('get', '/news/88/like')
  })
})

describe('news API - 收藏', () => {
  it('listFavorites → GET /news/favorites，params 透传', async () => {
    const params = { page: 1, page_size: 10 }
    await newsApi.listFavorites(params)
    expect(requestMock).toHaveBeenCalledWith('get', '/news/favorites', { params })
  })

  it('listFavorites 无参 → params undefined 仍可调用', async () => {
    await newsApi.listFavorites()
    expect(requestMock).toHaveBeenCalledWith('get', '/news/favorites', { params: undefined })
  })

  it('favNews → POST /news/:id/fav（toggle）', async () => {
    await newsApi.favNews(88)
    expect(requestMock).toHaveBeenCalledWith('post', '/news/88/fav')
  })

  it('getNewsFavStatus → GET /news/:id/fav', async () => {
    await newsApi.getNewsFavStatus(88)
    expect(requestMock).toHaveBeenCalledWith('get', '/news/88/fav')
  })

  it('getNewsFavStatus 不同 id 切换路径插值', async () => {
    await newsApi.getNewsFavStatus(1024)
    expect(requestMock).toHaveBeenCalledWith('get', '/news/1024/fav')
  })
})

describe('news API - 评论', () => {
  it('listComments → GET /news/:id/comments，params 透传', async () => {
    const params = { page: 1, page_size: 20 }
    await newsApi.listComments(88, params)
    expect(requestMock).toHaveBeenCalledWith('get', '/news/88/comments', { params })
  })

  it('listComments 无 params → params undefined 仍可调用', async () => {
    await newsApi.listComments(88)
    expect(requestMock).toHaveBeenCalledWith('get', '/news/88/comments', { params: undefined })
  })

  it('listComments 不同 newsId 切换路径插值', async () => {
    await newsApi.listComments(1024)
    expect(requestMock).toHaveBeenCalledWith('get', '/news/1024/comments', { params: undefined })
  })

  it('createComment → POST /news/:id/comments，body 透传', async () => {
    const data = { content: '好的', parent_id: 5, reply_to: '张三' }
    await newsApi.createComment(88, data)
    expect(requestMock).toHaveBeenCalledWith('post', '/news/88/comments', data)
  })

  it('createComment 仅含 content 也可', async () => {
    await newsApi.createComment(88, { content: '不错' })
    expect(requestMock).toHaveBeenCalledWith('post', '/news/88/comments', { content: '不错' })
  })

  it('deleteComment → DELETE /news/comments/:id', async () => {
    await newsApi.deleteComment(77)
    expect(requestMock).toHaveBeenCalledWith('delete', '/news/comments/77')
  })

  it('deleteComment 不同 commentId 切换路径插值', async () => {
    await newsApi.deleteComment(999)
    expect(requestMock).toHaveBeenCalledWith('delete', '/news/comments/999')
  })
})

describe('news API - 消息通知', () => {
  it('listMessages → GET /news/messages，params 透传', async () => {
    const params = { page: 1, page_size: 20 }
    await newsApi.listMessages(params)
    expect(requestMock).toHaveBeenCalledWith('get', '/news/messages', { params })
  })

  it('listMessages 无参 → params undefined 仍可调用', async () => {
    await newsApi.listMessages()
    expect(requestMock).toHaveBeenCalledWith('get', '/news/messages', { params: undefined })
  })

  it('getUnreadCount → GET /news/messages/unread 无参', async () => {
    await newsApi.getUnreadCount()
    expect(requestMock).toHaveBeenCalledWith('get', '/news/messages/unread')
  })

  it('markMessagesRead → PUT /news/messages/read，ids 透传', async () => {
    await newsApi.markMessagesRead([1, 2, 3])
    expect(requestMock).toHaveBeenCalledWith('put', '/news/messages/read', { ids: [1, 2, 3] })
  })

  it('markMessagesRead 空数组 → body ids 仍为数组', async () => {
    await newsApi.markMessagesRead([])
    expect(requestMock).toHaveBeenCalledWith('put', '/news/messages/read', { ids: [] })
  })

  it('markMessagesRead 不传参 → body ids 兜底为空数组', async () => {
    await newsApi.markMessagesRead()
    expect(requestMock).toHaveBeenCalledWith('put', '/news/messages/read', { ids: [] })
  })

  it('markMessagesRead 传 null → body ids 兜底为空数组', async () => {
    await newsApi.markMessagesRead(null)
    expect(requestMock).toHaveBeenCalledWith('put', '/news/messages/read', { ids: [] })
  })
})
