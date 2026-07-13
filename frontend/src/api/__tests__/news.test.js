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
