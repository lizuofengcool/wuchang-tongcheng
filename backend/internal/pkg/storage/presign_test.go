// Package storage 预签名直传（PresignPut / AccessURL）单元测试。
//
// MinIO 实现走 minio-go PresignedPutObject，仅在本地用 AK/SK 做 AWS SigV4 签名，
// 不发起任何网络请求，因此可在无 MinIO 服务的环境下离线验证 URL 拼装与签名。
// LocalStorage / QiniuStorage 不支持预签名直传，返回 ErrPresignNotSupported。
package storage

import (
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"wuchang-tongcheng/internal/pkg/config"
)

// newMinIOStorageForTest 直接构造 MinIOStorage（跳过 ensureBucket 的网络调用），
// 仅用于验证 PresignPut/AccessURL 的本地逻辑（签名/URL 拼装）。
func newMinIOStorageForTest(t *testing.T) *MinIOStorage {
	t.Helper()
	client, err := minio.New("localhost:9000", &minio.Options{
		Creds:  credentials.NewStaticV4("minioadmin", "minioadmin", ""),
		Secure: false,
		Region: "us-east-1",
	})
	require.NoError(t, err)
	return &MinIOStorage{
		client:   client,
		bucket:   "wctc-test",
		domain:   "http://localhost:9000",
		region:   "us-east-1",
		useSSL:   false,
		endpoint: "localhost:9000",
	}
}

// TestLocalStorage_PresignNotSupported 本地存储不支持预签名直传
func TestLocalStorage_PresignNotSupported(t *testing.T) {
	s, err := NewLocalStorage(&config.StorageConfig{Type: "local", Bucket: "/tmp/wctc-presign-local"})
	require.NoError(t, err)

	_, _, _, err = s.PresignPut("a.jpg", 15*time.Minute)
	assert.ErrorIs(t, err, ErrPresignNotSupported)

	_, err = s.AccessURL("2026/07/x.jpg")
	assert.ErrorIs(t, err, ErrPresignNotSupported)
}

// TestQiniuStorage_PresignNotSupported 七牛云暂不支持预签名直传
func TestQiniuStorage_PresignNotSupported(t *testing.T) {
	// 零值 QiniuStorage 即可：PresignPut/AccessURL 不依赖任何字段
	s := &QiniuStorage{}

	_, _, _, err := s.PresignPut("a.jpg", 15*time.Minute)
	assert.ErrorIs(t, err, ErrPresignNotSupported)

	_, err = s.AccessURL("2026/07/x.jpg")
	assert.ErrorIs(t, err, ErrPresignNotSupported)
}

// TestMinIOStorage_PresignPut 离线签名生成预签名 PUT URL
func TestMinIOStorage_PresignPut(t *testing.T) {
	s := newMinIOStorageForTest(t)

	uploadURL, objectName, accessURL, err := s.PresignPut("photo.JPG", 15*time.Minute)
	require.NoError(t, err)

	// objectName：按日期分目录 + 唯一文件名 + 原扩展名（保留大小写扩展名）
	require.NotEmpty(t, objectName)
	assert.True(t, strings.HasSuffix(objectName, ".JPG"), "对象名应保留原扩展名, got %s", objectName)
	assert.Contains(t, objectName, "/", "对象名应按日期分目录, got %s", objectName)

	// accessURL == {domain}/{bucket}/{objectName}
	assert.Equal(t, "http://localhost:9000/wctc-test/"+objectName, accessURL)

	// uploadURL 应可解析、指向 MinIO、携带 SigV4 签名参数
	u, err := url.Parse(uploadURL)
	require.NoError(t, err)
	assert.Equal(t, "localhost:9000", u.Host)
	assert.Equal(t, "/wctc-test/"+objectName, u.Path)
	q := u.Query()
	assert.NotEmpty(t, q.Get("X-Amz-Signature"), "应携带 AWS SigV4 签名")
	assert.NotEmpty(t, q.Get("X-Amz-Credential"))
	assert.Equal(t, "900", q.Get("X-Amz-Expires"), "15min=900s 应写入有效期")
}

// TestMinIOStorage_PresignPut_DefaultExpiry expiry<=0 时默认 15 分钟
func TestMinIOStorage_PresignPut_DefaultExpiry(t *testing.T) {
	s := newMinIOStorageForTest(t)

	uploadURL, _, _, err := s.PresignPut("a.png", 0)
	require.NoError(t, err)
	u, err := url.Parse(uploadURL)
	require.NoError(t, err)
	assert.Equal(t, "900", u.Query().Get("X-Amz-Expires"))
}

// TestMinIOStorage_PresignPut_ExpiryTooLarge 超出 S3 协议 7 天上限返回错误
func TestMinIOStorage_PresignPut_ExpiryTooLarge(t *testing.T) {
	s := newMinIOStorageForTest(t)

	_, _, _, err := s.PresignPut("a.png", 8*24*time.Hour)
	require.Error(t, err)
}

// TestMinIOStorage_AccessURL 按对象名拼装访问 URL
func TestMinIOStorage_AccessURL(t *testing.T) {
	s := newMinIOStorageForTest(t)

	got, err := s.AccessURL("2026/07/123.jpg")
	require.NoError(t, err)
	assert.Equal(t, "http://localhost:9000/wctc-test/2026/07/123.jpg", got)

	// 空对象名应拒绝
	_, err = s.AccessURL("")
	require.Error(t, err)
}

// TestMinIOStorage_PresignPut_UniqueObjectNames 连续两次生成的对象名应唯一
func TestMinIOStorage_PresignPut_UniqueObjectNames(t *testing.T) {
	s := newMinIOStorageForTest(t)

	_, o1, _, err := s.PresignPut("a.jpg", 15*time.Minute)
	require.NoError(t, err)
	_, o2, _, err := s.PresignPut("a.jpg", 15*time.Minute)
	require.NoError(t, err)
	assert.NotEqual(t, o1, o2, "连续生成的对象名应唯一")
}
