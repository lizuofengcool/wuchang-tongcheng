// 单元测试：src/api/setting.js
// 验证每个 API 函数调用 request 的方法、URL、入参（含 group 路径插值 + batchUpdateSettings 包裹 items）
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

import * as settingApi from '../setting'

beforeEach(() => {
  requestMock.mockReset()
  requestMock.mockImplementation(() => Promise.resolve({ code: 0, data: {} }))
})

describe('setting API', () => {
  it('getAllSettings → GET /setting（无参）', async () => {
    await settingApi.getAllSettings()
    expect(requestMock).toHaveBeenCalledWith('get', '/setting')
  })

  it('getSettingsByGroup → GET /setting/group/:group，路径插值', async () => {
    await settingApi.getSettingsByGroup('site')
    expect(requestMock).toHaveBeenCalledWith('get', '/setting/group/site')
  })

  it('getSettingsByGroup 不同 group → 路径切换', async () => {
    await settingApi.getSettingsByGroup('sms')
    expect(requestMock).toHaveBeenCalledWith('get', '/setting/group/sms')
  })

  it('createSetting → POST /setting，body 透传', async () => {
    const data = { key: 'site.name', value: '五常同城', group: 'site', type: 'string' }
    await settingApi.createSetting(data)
    expect(requestMock).toHaveBeenCalledWith('post', '/setting', data)
  })

  it('updateSetting → PUT /setting/:id，body 透传', async () => {
    const data = { value: '新名称' }
    await settingApi.updateSetting(11, data)
    expect(requestMock).toHaveBeenCalledWith('put', '/setting/11', data)
  })

  it('deleteSetting → DELETE /setting/:id', async () => {
    await settingApi.deleteSetting(11)
    expect(requestMock).toHaveBeenCalledWith('delete', '/setting/11')
  })

  it('batchUpdateSettings → PUT /setting/batch，items 包裹在 { items } 内', async () => {
    const items = [
      { id: 1, value: 'a' },
      { id: 2, value: 'b' }
    ]
    await settingApi.batchUpdateSettings(items)
    expect(requestMock).toHaveBeenCalledWith('put', '/setting/batch', { items })
  })

  it('batchUpdateSettings 空数组 → body 仍为 { items: [] }', async () => {
    await settingApi.batchUpdateSettings([])
    expect(requestMock).toHaveBeenCalledWith('put', '/setting/batch', { items: [] })
  })
})
