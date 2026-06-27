// Package storage 对象存储 - MinIO 实现
// 基于 S3 协议，兼容 AWS S3、阿里云 OSS、腾讯云 COS 等
package storage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"wuchang-tongcheng/internal/pkg/config"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// MinIOStorage MinIO 对象存储实现
type MinIOStorage struct {
	client   *minio.Client
	bucket   string // 存储桶名
	domain   string // 公开访问域名（用于拼接 URL）
	region   string // 区域
	useSSL   bool   // 是否启用 SSL
	endpoint string // 服务端点
}

// NewMinIOStorage 创建 MinIO 存储实例
//
// 配置约定：
//   - Domain: MinIO 服务端点（如 http://localhost:9000 或 s3.cn-east-1.amazonaws.com）
//   - AccessKey/SecretKey: 访问凭据
//   - Bucket: 存储桶名
//   - Region: 区域（S3 协议必填，本地 MinIO 可留空，会自动用 "us-east-1"）
func NewMinIOStorage(cfg *config.StorageConfig) (*MinIOStorage, error) {
	if cfg == nil {
		return nil, errors.New("minio storage config is nil")
	}
	endpoint := cfg.Domain
	if endpoint == "" {
		return nil, errors.New("minio endpoint(domain) is required")
	}

	// 解析端点，判断是否启用 SSL
	u, err := url.Parse(endpoint)
	if err != nil {
		return nil, fmt.Errorf("invalid minio endpoint %q: %w", endpoint, err)
	}
	useSSL := u.Scheme == "https"
	host := u.Host
	if host == "" {
		return nil, fmt.Errorf("invalid minio endpoint host: %s", endpoint)
	}

	region := cfg.Region
	if region == "" {
		region = "us-east-1" // S3 默认区域
	}

	// 创建 MinIO 客户端
	client, err := minio.New(host, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: useSSL,
		Region: region,
	})
	if err != nil {
		return nil, fmt.Errorf("create minio client failed: %w", err)
	}

	s := &MinIOStorage{
		client:   client,
		bucket:   cfg.Bucket,
		domain:   strings.TrimSuffix(endpoint, "/"),
		region:   region,
		useSSL:   useSSL,
		endpoint: host,
	}

	// 确保存储桶存在（幂等：不存在则创建并设置公开读策略）
	ctx := context.Background()
	if err := s.ensureBucket(ctx); err != nil {
		return nil, fmt.Errorf("ensure minio bucket failed: %w", err)
	}

	return s, nil
}

// ensureBucket 确保存储桶存在，不存在则创建
func (s *MinIOStorage) ensureBucket(ctx context.Context) error {
	if s.bucket == "" {
		return errors.New("minio bucket name is empty")
	}
	exists, err := s.client.BucketExists(ctx, s.bucket)
	if err != nil {
		return fmt.Errorf("check bucket exists failed: %w", err)
	}
	if !exists {
		if err := s.client.MakeBucket(ctx, s.bucket, minio.MakeBucketOptions{
			Region: s.region,
		}); err != nil {
			return fmt.Errorf("create bucket %q failed: %w", s.bucket, err)
		}
		// 设置匿名读策略，使上传的文件可通过 URL 公开访问
		policy := fmt.Sprintf(`{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Principal":{"AWS":["*"]},"Action":["s3:GetObject"],"Resource":["arn:aws:s3:::%s/*"]}]}`, s.bucket)
		if err := s.client.SetBucketPolicy(ctx, s.bucket, policy); err != nil {
			// 策略设置失败不阻塞，仅记录（生产环境应通过运维预配置）
			fmt.Printf("[minio] warn: set bucket policy failed: %v\n", err)
		}
	}
	return nil
}

// Save 保存文件，返回可访问的 URL
func (s *MinIOStorage) Save(filename string, reader io.Reader) (string, error) {
	// 按日期分目录：2026/06/xxx.jpg，避免单目录文件过多
	dateDir := time.Now().Format("2006/01")
	ext := filepath.Ext(filename)
	objectName := fmt.Sprintf("%s/%d%s", dateDir, time.Now().UnixNano(), ext)

	ctx := context.Background()
	// 获取文件大小（io.Reader 通常可 Seek，不能则用 -1 让 MinIO 自动分块）
	size, err := getReaderSize(reader)
	if err != nil {
		size = -1
	}

	_, err = s.client.PutObject(ctx, s.bucket, objectName, reader, size, minio.PutObjectOptions{
		ContentType: guessContentType(ext),
	})
	if err != nil {
		return "", fmt.Errorf("minio put object failed: %w", err)
	}

	// 拼接公开访问 URL：{domain}/{bucket}/{objectName}
	return fmt.Sprintf("%s/%s/%s", s.domain, s.bucket, objectName), nil
}

// Delete 删除文件
func (s *MinIOStorage) Delete(fileURL string) error {
	// 从 URL 中提取 objectName：{domain}/{bucket}/{objectName}
	prefix := fmt.Sprintf("%s/%s/", s.domain, s.bucket)
	if !strings.HasPrefix(fileURL, prefix) {
		return errors.New("invalid minio file url")
	}
	objectName := strings.TrimPrefix(fileURL, prefix)
	if objectName == "" {
		return errors.New("empty object name")
	}
	return s.client.RemoveObject(context.Background(), s.bucket, objectName, minio.RemoveObjectOptions{})
}

// getReaderSize 尝试获取 reader 的字节数
func getReaderSize(r io.Reader) (int64, error) {
	type sizer interface{ Size() (int64, error) }
	type seeker interface{ Seek(offset int64, whence int) (int64, error) }
	if s, ok := r.(sizer); ok {
		return s.Size()
	}
	if sk, ok := r.(seeker); ok {
		pos, err := sk.Seek(0, io.SeekCurrent)
		if err != nil {
			return -1, err
		}
		end, err := sk.Seek(0, io.SeekEnd)
		if err != nil {
			return -1, err
		}
		_, _ = sk.Seek(pos, io.SeekStart)
		return end, nil
	}
	return -1, nil
}

// guessContentType 根据扩展名猜测 Content-Type
func guessContentType(ext string) string {
	ext = strings.ToLower(strings.TrimPrefix(ext, "."))
	switch ext {
	case ".jpg", "jpg", "jpeg":
		return "image/jpeg"
	case ".png", "png":
		return "image/png"
	case ".gif", "gif":
		return "image/gif"
	case ".webp", "webp":
		return "image/webp"
	case ".svg", "svg":
		return "image/svg+xml"
	case ".pdf", "pdf":
		return "application/pdf"
	case ".mp4", "mp4":
		return "video/mp4"
	case ".mp3", "mp3":
		return "audio/mpeg"
	case ".zip", "zip":
		return "application/zip"
	case ".json", "json":
		return "application/json"
	case ".html", "html", "htm":
		return "text/html"
	case ".css", "css":
		return "text/css"
	case ".js", "js":
		return "application/javascript"
	default:
		return "application/octet-stream"
	}
}
