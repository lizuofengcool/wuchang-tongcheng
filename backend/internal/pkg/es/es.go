// Package es Elasticsearch 封装
// 基于 elastic/go-elasticsearch/v8 的客户端封装
// 与 redis/mq 包风格一致：包级单例 + Init() + GetClient() + IsAvailable() + Close()
// 设计要点：
//   - 未配置或连不上不阻塞服务启动，业务侧通过 IsAvailable() 优雅降级
//   - 提供文档级 CRUD（IndexDoc/DeleteDoc）和检索（SearchByQuery）
//   - 提供索引管理（CreateIndexIfNotExists）
//   - 使用 esapi 函数式风格（XxxRequest{}.Do(ctx, transport)），兼容 v8 客户端
package es

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"sync"
	"time"

	"wuchang-tongcheng/internal/pkg/config"
	"wuchang-tongcheng/internal/pkg/logger"

	es "github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"go.uber.org/zap"
)

var (
	mu     sync.RWMutex
	client *es.Client
)

// Init 初始化 Elasticsearch 客户端
// cfg 为 nil 或无 Addresses 时跳过（允许无 ES 降级运行）
func Init(cfg *config.ESConfig) error {
	if cfg == nil || len(cfg.Addresses) == 0 {
		logger.Info("Elasticsearch 未配置，跳过初始化（业务可降级）")
		return nil
	}

	cli, err := es.NewClient(es.Config{
		Addresses: cfg.Addresses,
		Username:  cfg.Username,
		Password:  cfg.Password,
	})
	if err != nil {
		return fmt.Errorf("create es client failed: %w", err)
	}

	// 测试连接（5 秒超时）
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	resp, err := esapi.PingRequest{}.Do(ctx, cli)
	if err != nil {
		return fmt.Errorf("ping es failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.IsError() {
		return fmt.Errorf("ping es failed: %s", resp.String())
	}

	mu.Lock()
	client = cli
	mu.Unlock()

	logger.Info("Elasticsearch 初始化成功", zap.Strings("addrs", cfg.Addresses))
	return nil
}

// GetClient 获取 Elasticsearch 客户端
func GetClient() *es.Client {
	mu.RLock()
	defer mu.RUnlock()
	return client
}

// IsAvailable 检查 Elasticsearch 是否已初始化
func IsAvailable() bool {
	mu.RLock()
	defer mu.RUnlock()
	return client != nil
}

// Close 关闭客户端（ES v8 客户端无显式 Close，保留方法以对齐 redis/mq 风格）
func Close() error {
	mu.Lock()
	client = nil
	mu.Unlock()
	return nil
}

// IndexDoc 索引（创建或全量更新）一个文档
//   - index: 索引名
//   - id: 文档 ID（必须唯一，建议用业务主键）
//   - doc: 任意可 JSON 序列化的对象
func IndexDoc(ctx context.Context, index, id string, doc interface{}) error {
	c := GetClient()
	if c == nil {
		return ErrNotAvailable
	}
	body, err := json.Marshal(doc)
	if err != nil {
		return fmt.Errorf("marshal doc failed: %w", err)
	}
	resp, err := esapi.IndexRequest{
		Index:      index,
		DocumentID: id,
		Body:       bytes.NewReader(body),
		Refresh:    "true",
	}.Do(ctx, c)
	if err != nil {
		return fmt.Errorf("index doc failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.IsError() {
		return fmt.Errorf("index doc failed: %s", resp.String())
	}
	return nil
}

// DeleteDoc 按 ID 删除文档（不存在不视为错误）
func DeleteDoc(ctx context.Context, index, id string) error {
	c := GetClient()
	if c == nil {
		return ErrNotAvailable
	}
	resp, err := esapi.DeleteRequest{
		Index:      index,
		DocumentID: id,
		Refresh:    "true",
	}.Do(ctx, c)
	if err != nil {
		// ES v8 在文档不存在时返回 404，但客户端可能包装成错误
		if strings.Contains(err.Error(), "404") || strings.Contains(strings.ToLower(err.Error()), "not_found") {
			return nil
		}
		return fmt.Errorf("delete doc failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.IsError() {
		// 404 不视为错误
		if resp.StatusCode == 404 {
			return nil
		}
		return fmt.Errorf("delete doc failed: %s", resp.String())
	}
	return nil
}

// SearchResult 检索结果
type SearchResult struct {
	Total int64                    `json:"total"` // 命中总数
	Hits  []map[string]interface{} `json:"hits"`  // 命中文档列表（含 _source）
}

// SearchByQuery 使用 query DSL 进行检索
//   - index: 索引名
//   - query: ES query DSL 的 JSON 字符串，例如 {"query":{"match":{"title":"五常"}}}
//   - from/size: 分页
func SearchByQuery(ctx context.Context, index, query string, from, size int) (*SearchResult, error) {
	c := GetClient()
	if c == nil {
		return nil, ErrNotAvailable
	}
	if size <= 0 {
		size = 10
	}
	if size > 100 {
		size = 100
	}
	if from < 0 {
		from = 0
	}
	resp, err := esapi.SearchRequest{
		Index: []string{index},
		Body: strings.NewReader(query),
		From: &from,
		Size: &size,
	}.Do(ctx, c)
	if err != nil {
		return nil, fmt.Errorf("search failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.IsError() {
		return nil, fmt.Errorf("search failed: %s", resp.String())
	}

	// 解析响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read search response failed: %w", err)
	}
	var raw struct {
		Hits struct {
			Total struct {
				Value int64 `json:"value"`
			} `json:"total"`
			Hits []struct {
				Source map[string]interface{} `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("unmarshal search response failed: %w", err)
	}

	result := &SearchResult{
		Total: raw.Hits.Total.Value,
		Hits:  make([]map[string]interface{}, 0, len(raw.Hits.Hits)),
	}
	for _, h := range raw.Hits.Hits {
		if h.Source != nil {
			result.Hits = append(result.Hits, h.Source)
		}
	}
	return result, nil
}

// CreateIndexIfNotExists 索引不存在则创建（带 mappings）
// mappings 为 ES mappings JSON 字符串；为空则创建空 mappings 索引
func CreateIndexIfNotExists(ctx context.Context, index, mappings string) error {
	c := GetClient()
	if c == nil {
		return ErrNotAvailable
	}
	// 检查是否已存在
	existsResp, err := esapi.IndicesExistsRequest{
		Index: []string{index},
	}.Do(ctx, c)
	if err != nil {
		return fmt.Errorf("check index exists failed: %w", err)
	}
	defer existsResp.Body.Close()
	if existsResp.StatusCode == 200 {
		return nil // 已存在
	}

	// 创建索引（IndicesCreate 对应 PUT /{index}）
	req := esapi.IndicesCreateRequest{Index: index}
	if mappings != "" {
		req.Body = bytes.NewReader([]byte(mappings))
	}
	resp, err := req.Do(ctx, c)
	if err != nil {
		return fmt.Errorf("create index failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.IsError() {
		return fmt.Errorf("create index failed: %s", resp.String())
	}
	logger.Info("Elasticsearch 索引已创建", zap.String("index", index))
	return nil
}

// ErrNotAvailable Elasticsearch 未初始化或不可用
var ErrNotAvailable = errors.New("elasticsearch not available")

// （以下为可选的便捷 helper，用于业务侧避免反复 strconv）

// IDToStr uint 主键转字符串（ES 文档 ID 必须为字符串）
func IDToStr(id uint) string {
	return strconv.FormatUint(uint64(id), 10)
}
