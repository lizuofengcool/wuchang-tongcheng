// Package service 预签名直传业务逻辑单元测试。
//
// 覆盖 PresignUpload / CommitUpload 的输入校验路径（类型/大小），
// 以及当前端使用不支持预签名的本地存储（GetStorage 兜底 LocalStorage）时
// 返回 storage.ErrPresignNotSupported 的降级路径。
// 这些路径不依赖 DB / 真实对象存储，可在无基础设施环境离线运行。
package service

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"wuchang-tongcheng/internal/pkg/storage"
)

// maxFileSize 与 service 包常量一致（50MB），此处仅用于断言
const testMaxFileSize = 50 * 1024 * 1024

// TestPresignUpload_InvalidType 不支持的扩展名应拒绝
func TestPresignUpload_InvalidType(t *testing.T) {
	svc := NewFileService(nil)

	_, err := svc.PresignUpload(1, 100, "malware.exe")
	assert.ErrorIs(t, err, ErrFileTypeInvalid)
}

// TestPresignUpload_LocalStorageNotSupported 本地存储（GetStorage 兜底）不支持预签名直传
func TestPresignUpload_LocalStorageNotSupported(t *testing.T) {
	svc := NewFileService(nil)

	_, err := svc.PresignUpload(1, 100, "photo.jpg")
	assert.ErrorIs(t, err, storage.ErrPresignNotSupported)
}

// TestCommitUpload_EmptySize 空文件应拒绝
func TestCommitUpload_EmptySize(t *testing.T) {
	svc := NewFileService(nil)

	_, err := svc.CommitUpload(1, 100, "photo.jpg", "2026/07/x.jpg", "image/jpeg", 0)
	assert.ErrorIs(t, err, ErrFileEmpty)
}

// TestCommitUpload_NegativeSize 负大小视为空文件
func TestCommitUpload_NegativeSize(t *testing.T) {
	svc := NewFileService(nil)

	_, err := svc.CommitUpload(1, 100, "photo.jpg", "2026/07/x.jpg", "image/jpeg", -1)
	assert.ErrorIs(t, err, ErrFileEmpty)
}

// TestCommitUpload_TooLarge 超出 50MB 上限应拒绝
func TestCommitUpload_TooLarge(t *testing.T) {
	svc := NewFileService(nil)

	_, err := svc.CommitUpload(1, 100, "big.mp4", "2026/07/x.mp4", "video/mp4", testMaxFileSize+1)
	assert.ErrorIs(t, err, ErrFileTooLarge)
}

// TestCommitUpload_InvalidType 不支持的扩展名应拒绝
func TestCommitUpload_InvalidType(t *testing.T) {
	svc := NewFileService(nil)

	_, err := svc.CommitUpload(1, 100, "a.unknownext", "2026/07/x.unknownext", "", 1024)
	assert.ErrorIs(t, err, ErrFileTypeInvalid)
}

// TestCommitUpload_LocalStorageNotSupported 本地存储不支持按对象名提交记录
func TestCommitUpload_LocalStorageNotSupported(t *testing.T) {
	svc := NewFileService(nil)

	_, err := svc.CommitUpload(1, 100, "photo.jpg", "2026/07/x.jpg", "image/jpeg", 1024)
	assert.ErrorIs(t, err, storage.ErrPresignNotSupported)
}

// TestGuessMIMEByExt 扩展名到 MIME 推断
func TestGuessMIMEByExt(t *testing.T) {
	cases := map[string]string{
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".png":  "image/png",
		".gif":  "image/gif",
		".webp": "image/webp",
		".mp4":  "video/mp4",
		".mov":  "video/quicktime",
		".mp3":  "audio/mpeg",
		".wav":  "audio/wav",
		".pdf":  "application/pdf",
		".zip":  "application/zip",
		".txt":  "text/plain",
		".xyz":  "application/octet-stream", // 未知扩展名兜底
		"":      "application/octet-stream", // 无扩展名兜底
	}
	for ext, want := range cases {
		assert.Equal(t, want, guessMIMEByExt(ext), "guessMIMEByExt(%q)", ext)
	}
}

// TestPresignExpiryConstant 有效期常量应为 15 分钟（与 DTO expires_in 字段一致）
func TestPresignExpiryConstant(t *testing.T) {
	assert.Equal(t, 15*time.Minute, presignExpiry)
	assert.Equal(t, 900, int(presignExpiry.Seconds()))
}
