// 单元测试：src/api/file.js
// 验证：上传文件构造 FormData + multipart 头 + onUploadProgress 回调；
//      listFiles / deleteFile / getSTSCredentials 调用 request 的方法/URL/入参
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

import * as fileApi from '../file'

beforeEach(() => {
  requestMock.mockReset()
  requestMock.mockImplementation(() => Promise.resolve({ code: 0, data: {} }))
})

describe('file API - 上传', () => {
  it('uploadFile → POST /file/upload，FormData 携带 file，multipart 头 + onUploadProgress 配置', async () => {
    await fileApi.uploadFile('fake-file-blob')
    expect(requestMock).toHaveBeenCalledTimes(1)
    const [, url, formData, config] = requestMock.mock.calls[0]
    expect(url).toBe('/file/upload')
    expect(formData).toBeInstanceOf(FormData)
    expect(formData.get('file')).toBe('fake-file-blob')
    expect(config.headers['Content-Type']).toBe('multipart/form-data')
    expect(typeof config.onUploadProgress).toBe('function')
  })

  it('uploadFile onProgress 回调 → onUploadProgress 内部按 loaded/total 换算百分比', async () => {
    const onProgress = vi.fn()
    await fileApi.uploadFile('blob', onProgress)
    const config = requestMock.mock.calls[0][3]
    // e.total 存在 → 计算 Math.round(loaded*100/total)
    config.onUploadProgress({ loaded: 50, total: 200 })
    expect(onProgress).toHaveBeenCalledWith(25)
    // 再次触发 → 累计调用两次
    config.onUploadProgress({ loaded: 200, total: 200 })
    expect(onProgress).toHaveBeenCalledWith(100)
    expect(onProgress).toHaveBeenCalledTimes(2)
  })

  it('uploadFile onProgress 回调 → e.total 缺失时不调用 onProgress（避免除零）', async () => {
    const onProgress = vi.fn()
    await fileApi.uploadFile('blob', onProgress)
    const config = requestMock.mock.calls[0][3]
    config.onUploadProgress({ loaded: 50, total: 0 })
    expect(onProgress).not.toHaveBeenCalled()
  })

  it('uploadFile 未传 onProgress → onUploadProgress 仍可被调用且不抛错', async () => {
    await fileApi.uploadFile('blob')
    const config = requestMock.mock.calls[0][3]
    expect(() => config.onUploadProgress({ loaded: 1, total: 100 })).not.toThrow()
  })
})

describe('file API - 列表/删除/STS', () => {
  it('listFiles → GET /file，params 透传', async () => {
    const params = { page: 1, page_size: 20, type: 'image' }
    await fileApi.listFiles(params)
    expect(requestMock).toHaveBeenCalledWith('get', '/file', { params })
  })

  it('listFiles 无参 → params undefined', async () => {
    await fileApi.listFiles()
    expect(requestMock).toHaveBeenCalledWith('get', '/file', { params: undefined })
  })

  it('deleteFile → DELETE /file/:id', async () => {
    await fileApi.deleteFile(33)
    expect(requestMock).toHaveBeenCalledWith('delete', '/file/33')
  })

  it('getSTSCredentials → POST /file/sts（无 body）', async () => {
    await fileApi.getSTSCredentials()
    expect(requestMock).toHaveBeenCalledWith('post', '/file/sts')
  })
})
