// 单元测试：src/api/region.js
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

import * as regionApi from '../region'

beforeEach(() => {
  requestMock.mockReset()
  requestMock.mockImplementation(() => Promise.resolve({ code: 0, data: {} }))
})

describe('region API', () => {
  it('getRegionTree → GET /region/tree（无参）', async () => {
    await regionApi.getRegionTree()
    expect(requestMock).toHaveBeenCalledWith('get', '/region/tree')
  })

  it('getRegionChildren → GET /region/children，parent_id 走 query', async () => {
    await regionApi.getRegionChildren(2)
    expect(requestMock).toHaveBeenCalledWith('get', '/region/children', {
      params: { parent_id: 2 }
    })
  })

  it('getRegionChildren parent_id=0 → 顶层地区', async () => {
    await regionApi.getRegionChildren(0)
    expect(requestMock).toHaveBeenCalledWith('get', '/region/children', {
      params: { parent_id: 0 }
    })
  })

  it('getRegion → GET /region/:id', async () => {
    await regionApi.getRegion(5)
    expect(requestMock).toHaveBeenCalledWith('get', '/region/5')
  })

  it('createRegion → POST /region，body 透传', async () => {
    const data = { name: '武昌区', parent_id: 2, code: '420106' }
    await regionApi.createRegion(data)
    expect(requestMock).toHaveBeenCalledWith('post', '/region', data)
  })

  it('updateRegion → PUT /region/:id，body 透传', async () => {
    const data = { name: '新武昌区' }
    await regionApi.updateRegion(5, data)
    expect(requestMock).toHaveBeenCalledWith('put', '/region/5', data)
  })

  it('deleteRegion → DELETE /region/:id', async () => {
    await regionApi.deleteRegion(5)
    expect(requestMock).toHaveBeenCalledWith('delete', '/region/5')
  })
})
