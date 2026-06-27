// Package indexer 头条索引器
// 封装"DB 写入 -> 索引同步"的解耦逻辑：
//   - MQIndexer：MQ 可用时，发布事件到队列，由消费者异步写 ES（推荐）
//   - DirectESIndexer：MQ 不可用但 ES 可用时，直接异步写 ES（降级）
//   - NoopIndexer：MQ 和 ES 都不可用时，跳过
// 通过 New() 工厂方法按可用性自动选择实现
package indexer

import (
	"context"
	"encoding/json"
	"time"

	"wuchang-tongcheng/internal/modules/news/model"
	"wuchang-tongcheng/internal/modules/news/repository"
	"wuchang-tongcheng/internal/pkg/es"
	"wuchang-tongcheng/internal/pkg/logger"
	"wuchang-tongcheng/internal/pkg/mq"

	"go.uber.org/zap"
)

const (
	// QueueName MQ 队列名
	QueueName = "news.es.index"
	// IndexName ES 索引名
	IndexName = "news"

	// IndexMappings ES 索引 mappings
	// 默认用 standard 分词器（ES 内置，开箱即用）
	// 如已安装 ik 分词器插件（analysis-ik），可将 analyzer 改为 "ik_max_word"、
	// search_analyzer 改为 "ik_smart" 以获得更好的中文分词效果
	IndexMappings = `{
  "mappings": {
    "properties": {
      "id": {"type": "long"},
      "title": {"type": "text", "analyzer": "standard"},
      "content": {"type": "text", "analyzer": "standard"},
      "summary": {"type": "text", "analyzer": "standard"},
      "author_id": {"type": "long"},
      "author_name": {"type": "keyword"},
      "category_id": {"type": "long"},
      "tags": {"type": "text", "analyzer": "standard"},
      "status": {"type": "integer"},
      "region_id": {"type": "long"},
      "view_count": {"type": "integer"},
      "like_count": {"type": "integer"},
      "published_at": {"type": "date"},
      "created_at": {"type": "date"}
    }
  }
}`
)

// NewsIndexMessage MQ 消息体
type NewsIndexMessage struct {
	Action string `json:"action"` // "index" 或 "delete"
	NewsID uint   `json:"news_id"`
}

// Indexer 索引器接口
type Indexer interface {
	// OnIndex 触发索引更新（fire-and-forget，不阻塞主流程）
	OnIndex(news *model.News)
	// OnDelete 触发索引删除
	OnDelete(newsID uint)
}

// NoopIndexer 空实现（MQ 和 ES 都不可用时）
type NoopIndexer struct{}

func (NoopIndexer) OnIndex(*model.News) {}
func (NoopIndexer) OnDelete(uint)       {}

// MQIndexer 通过 MQ 异步触发 ES 索引（推荐方案）
type MQIndexer struct{}

func (MQIndexer) OnIndex(news *model.News) {
	if news == nil || !mq.IsAvailable() {
		return
	}
	publishMessage(NewsIndexMessage{Action: "index", NewsID: news.ID})
}

func (MQIndexer) OnDelete(newsID uint) {
	if !mq.IsAvailable() {
		return
	}
	publishMessage(NewsIndexMessage{Action: "delete", NewsID: newsID})
}

func publishMessage(msg NewsIndexMessage) {
	body, err := json.Marshal(msg)
	if err != nil {
		logger.Warn("序列化索引消息失败", zap.Error(err))
		return
	}
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		if err := mq.GetClient().SimplePublish(ctx, QueueName, body); err != nil {
			logger.Warn("发送 news 索引消息失败",
				zap.String("action", msg.Action),
				zap.Uint("news_id", msg.NewsID),
				zap.Error(err))
		}
	}()
}

// DirectESIndexer 直接同步调 ES（MQ 不可用但 ES 可用时的降级方案）
type DirectESIndexer struct{}

func (DirectESIndexer) OnIndex(news *model.News) {
	if news == nil || !es.IsAvailable() {
		return
	}
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := es.IndexDoc(ctx, IndexName, es.IDToStr(news.ID), news); err != nil {
			logger.Warn("ES 索引失败",
				zap.Uint("news_id", news.ID),
				zap.Error(err))
		}
	}()
}

func (DirectESIndexer) OnDelete(newsID uint) {
	if !es.IsAvailable() {
		return
	}
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := es.DeleteDoc(ctx, IndexName, es.IDToStr(newsID)); err != nil {
			logger.Warn("ES 删除失败",
				zap.Uint("news_id", newsID),
				zap.Error(err))
		}
	}()
}

// New 按可用性自动选择 Indexer 实现
func New() Indexer {
	if mq.IsAvailable() {
		return MQIndexer{}
	}
	if es.IsAvailable() {
		return DirectESIndexer{}
	}
	return NoopIndexer{}
}

// StartConsumer 启动 MQ 消费者（仅 MQIndexer 场景需要调用）
// 消费 "news.es.index" 队列，根据消息 action 调用 ES 索引/删除
// newsRepo 用于 action=index 时从 DB 查最新 news 数据
func StartConsumer(ctx context.Context, newsRepo repository.NewsRepository) error {
	if !mq.IsAvailable() {
		return nil
	}
	// 确保 ES 索引存在（仅当 ES 可用时）
	if es.IsAvailable() {
		if err := es.CreateIndexIfNotExists(ctx, IndexName, IndexMappings); err != nil {
			logger.Warn("创建 ES 索引失败，消费者仍会启动", zap.Error(err))
		}
	}
	return mq.GetClient().Consume(ctx, QueueName, func(body []byte) error {
		var msg NewsIndexMessage
		if err := json.Unmarshal(body, &msg); err != nil {
			return err
		}
		cctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		switch msg.Action {
		case "index":
			news, err := newsRepo.FindByID(msg.NewsID)
			if err != nil {
				return err
			}
			if !es.IsAvailable() {
				return nil // ES 不可用，跳过（消息已 ack，避免无限重试）
			}
			return es.IndexDoc(cctx, IndexName, es.IDToStr(news.ID), news)
		case "delete":
			if !es.IsAvailable() {
				return nil
			}
			return es.DeleteDoc(cctx, IndexName, es.IDToStr(msg.NewsID))
		}
		return nil
	})
}
