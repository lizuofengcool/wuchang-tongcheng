// 单元测试：src/api/category.js
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

import * as categoryApi from '../category'

beforeEach(() => {
  requestMock.mockReset()
  requestMock.mockImplementation(() => Promise.resolve({ code: 0, data: {} }))
})

describe('category API', () => {
  it('getCategoryTree → GET /category/tree（无参）', async () => {
    await categoryApi.getCategoryTree()
    expect(requestMock).toHaveBeenCalledWith('get', '/category/tree')
  })

  it('getCategoryChildren → GET /category/children，parent_id 走 query', async () => {
    await categoryApi.getCategoryChildren(5)
    expect(requestMock).toHaveBeenCalledWith('get', '/category/children', {
      params: { parent_id: 5 }
    })
  })

  it('getCategoryChildren 不同 parent_id → query 切换', async () => {
    await categoryApi.getCategoryChildren(0)
    expect(requestMock).toHaveBeenCalledWith('get', '/category/children', {
      params: { parent_id: 0 }
    })
  })

  it('getCategory → GET /category/:id，路径插值', async () => {
    await categoryApi.getCategory(12)
    expect(requestMock).toHaveBeenCalledWith('get', '/category/12')
  })

  it('createCategory → POST /category，body 透传', async () => {
    const data = { name: '分类A', parent_id: 1, sort: 10 }
    await categoryApi.createCategory(data)
    expect(requestMock).toHaveBeenCalledWith('post', '/category', data)
  })

  it('updateCategory → PUT /category/:id，body 透传', async () => {
    const data = { name: '分类B' }
    await categoryApi.updateCategory(12, data)
    expect(requestMock).toHaveBeenCalledWith('put', '/category/12', data)
  })

  it('deleteCategory → DELETE /category/:id', async () => {
    await categoryApi.deleteCategory(12)
    expect(requestMock).toHaveBeenCalledWith('delete', '/category/12')
  })
})
