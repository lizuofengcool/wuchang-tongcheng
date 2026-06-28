// Package storage 对象存储 - 七牛云 Kodo 实现
//
// 基于官方 SDK github.com/qiniu/go-sdk/v7（经典 storage 包）。
// 配置约定（复用 config.StorageConfig）：
//   - AccessKey/SecretKey: 七牛 AK/SK
//   - Bucket: 存储空间名
//   - Domain: 绑定的 CDN/源站域名（如 http://xxx.qiniudn.com 或 https://cdn.example.com）
//   - Region: 区域选择辅助（留空时使用默认 z0 华东）
//
// key 未配置（占位值/空值）时 NewQiniuStorage 返回错误，
// storage.Init 调用方据此降级到 local（与 amap 模式一致）。
package storage

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	"wuchang-tongcheng/internal/pkg/config"

	"github.com/qiniu/go-sdk/v7/auth"
	qiniustorage "github.com/qiniu/go-sdk/v7/storage"
)

// 占位凭据：Init 时排除这些值，避免误激活（与 amap.Init 一致）
var qiniuPlaceholderKeys = map[string]bool{
	"":                true,
	"your-access-key": true,
	"your_access_key": true,
	"your-secret-key": true,
	"your_secret_key": true,
	"your-qiniu-ak":   true,
	"your-qiniu-sk":   true,
}

// QiniuStorage 七牛云 Kodo 对象存储实现
type QiniuStorage struct {
	mac           *auth.Credentials          // 鉴权凭据
	bucket        string                     // 存储空间名
	domain        string                     // 绑定域名（拼接 URL 用，末尾不带斜杠）
	cfg           *qiniustorage.Config       // 上传配置
	bucketManager *qiniustorage.BucketManager // 空间管理器（删除用）
}

// NewQiniuStorage 创建七牛云存储实例
//
// 配置校验：AccessKey/SecretKey/Bucket/Domain 任一为空或占位值则返回错误，
// 调用方（storage.Init）据此降级到 local。
func NewQiniuStorage(cfg *config.StorageConfig) (*QiniuStorage, error) {
	if cfg == nil {
		return nil, errors.New("qiniu storage config is nil")
	}
	if qiniuPlaceholderKeys[cfg.AccessKey] {
		return nil, errors.New("qiniu access_key not configured (placeholder value)")
	}
	if qiniuPlaceholderKeys[cfg.SecretKey] {
		return nil, errors.New("qiniu secret_key not configured (placeholder value)")
	}
	if cfg.Bucket == "" {
		return nil, errors.New("qiniu bucket name is required")
	}
	if cfg.Domain == "" {
		return nil, errors.New("qiniu domain is required (CDN or origin domain)")
	}

	mac := auth.New(cfg.AccessKey, cfg.SecretKey)
	domain := strings.TrimSuffix(cfg.Domain, "/")

	// 上传配置：使用 HTTPS 域名时启用 HTTPS
	uploadCfg := &qiniustorage.Config{
		UseHTTPS: strings.HasPrefix(cfg.Domain, "https"),
	}

	// BucketManager 用于删除文件
	bm := qiniustorage.NewBucketManager(mac, uploadCfg)

	return &QiniuStorage{
		mac:           mac,
		bucket:        cfg.Bucket,
		domain:        domain,
		cfg:           uploadCfg,
		bucketManager: bm,
	}, nil
}

// Save 上传文件，返回可公开访问的 URL
//
// 按日期分目录：2026/06/{unixNano}{ext}，避免单目录文件过多。
// 使用 io.ReadAll 读取全部内容后通过 FormUploader 上传（FormUploader 适合小文件）。
func (s *QiniuStorage) Save(filename string, reader io.Reader) (string, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return "", fmt.Errorf("qiniu read file content failed: %w", err)
	}

	// 生成对象名：按日期分目录 + 唯一文件名 + 原扩展名
	dateDir := time.Now().Format("2006/01")
	ext := filepath.Ext(filename)
	objectName := fmt.Sprintf("%s/%d%s", dateDir, time.Now().UnixNano(), ext)

	// 生成上传凭证
	putPolicy := qiniustorage.PutPolicy{
		Scope: s.bucket,
	}
	upToken := putPolicy.UploadToken(s.mac)

	// 上传字节数据（FormUploader 适合小文件，大文件应用 ResumeUploader）
	ctx := context.Background()
	formUploader := qiniustorage.NewFormUploader(s.cfg)
	err = formUploader.Put(ctx, nil, objectName, upToken, bytes.NewReader(data), int64(len(data)),
		&qiniustorage.PutExtra{MimeType: guessContentType(ext)})
	if err != nil {
		return "", fmt.Errorf("qiniu upload failed: %w", err)
	}

	// 拼接公开访问 URL：{domain}/{objectName}
	return fmt.Sprintf("%s/%s", s.domain, objectName), nil
}

// Delete 删除文件
//
// 从 URL 中提取 objectName（去掉 domain 前缀），调用 BucketManager.Delete。
// 文件不存在视为成功（幂等）。
func (s *QiniuStorage) Delete(fileURL string) error {
	prefix := s.domain + "/"
	if !strings.HasPrefix(fileURL, prefix) {
		return errors.New("invalid qiniu file url")
	}
	objectName := strings.TrimPrefix(fileURL, prefix)
	if objectName == "" {
		return errors.New("empty qiniu object name")
	}

	err := s.bucketManager.Delete(s.bucket, objectName)
	if err != nil {
		// 幂等：612 表示文件不存在，视为成功
		errStr := err.Error()
		if strings.Contains(errStr, "no such file") || strings.Contains(errStr, "612") {
			return nil
		}
		return fmt.Errorf("qiniu delete failed: %w", err)
	}
	return nil
}
