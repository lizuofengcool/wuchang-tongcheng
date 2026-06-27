// Package mq RabbitMQ 封装
// 基于 amqp091-go 的连接/通道封装，提供简单的发布-订阅与点对点消息能力
// 与 redis 包风格一致：包级单例 + Init() + GetClient() + IsAvailable() + Close()
// 设计要点：
//   - 未配置或连不上不阻塞服务启动，业务侧通过 IsAvailable() 优雅降级
//   - 提供两种消息模型：
//   - SimplePublish(queue, body) 使用默认交换机直接路由到队列（点对点）
//   - Publish(exchange, routingKey, body) 使用 topic 交换机（发布-订阅）
package mq

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"wuchang-tongcheng/internal/pkg/config"
	"wuchang-tongcheng/internal/pkg/logger"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

var (
	mu     sync.RWMutex
	client *Client
)

// Client RabbitMQ 客户端封装（单连接 + 单共享通道）
type Client struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	url     string

	// 已声明的队列/交换机缓存，避免重复声明
	declaredQueues     map[string]struct{}
	declaredExchanges map[string]struct{}
	declaredMu        sync.Mutex
}

// Init 初始化 RabbitMQ 连接
// cfg 为 nil 或 Host 为空时跳过（允许无 MQ 降级运行）
func Init(cfg *config.RabbitMQConfig) error {
	if cfg == nil || cfg.Host == "" {
		logger.Info("RabbitMQ 未配置，跳过初始化（业务可降级）")
		return nil
	}

	url := cfg.GetURL()
	conn, err := amqp.Dial(url)
	if err != nil {
		return fmt.Errorf("connect rabbitmq failed: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return fmt.Errorf("open rabbitmq channel failed: %w", err)
	}

	// QoS：每个消费者未 ack 消息上限，防止某个慢消费者堆积
	if err := ch.Qos(16, 0, false); err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return fmt.Errorf("set qos failed: %w", err)
	}

	c := &Client{
		conn:              conn,
		channel:           ch,
		url:               url,
		declaredQueues:    make(map[string]struct{}),
		declaredExchanges: make(map[string]struct{}),
	}

	// 监听连接关闭
	go func() {
		errCh := make(chan *amqp.Error, 1)
		c.conn.NotifyClose(errCh)
		if e, ok := <-errCh; ok {
			mu.Lock()
			if client == c {
				client = nil
			}
			mu.Unlock()
			logger.Warn("RabbitMQ 连接已关闭", zap.Error(e))
		}
	}()

	mu.Lock()
	client = c
	mu.Unlock()

	logger.Info("RabbitMQ 初始化成功", zap.String("addr", fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)))
	return nil
}

// GetClient 获取 RabbitMQ 客户端
func GetClient() *Client {
	mu.RLock()
	defer mu.RUnlock()
	return client
}

// IsAvailable 检查 RabbitMQ 是否已初始化且连接未关闭
func IsAvailable() bool {
	mu.RLock()
	c := client
	mu.RUnlock()
	if c == nil {
		return false
	}
	return !c.conn.IsClosed()
}

// Close 关闭连接
func Close() error {
	mu.Lock()
	c := client
	client = nil
	mu.Unlock()
	if c == nil {
		return nil
	}
	var firstErr error
	if err := c.channel.Close(); err != nil && firstErr == nil {
		firstErr = err
	}
	if err := c.conn.Close(); err != nil && firstErr == nil {
		firstErr = err
	}
	return firstErr
}

// DeclareQueue 声明持久化队列（幂等，重复声明同名队列会被缓存跳过）
func (c *Client) DeclareQueue(name string) error {
	c.declaredMu.Lock()
	defer c.declaredMu.Unlock()

	if _, ok := c.declaredQueues[name]; ok {
		return nil
	}
	if _, err := c.channel.QueueDeclare(
		name,
		true,  // durable
		false, // auto-deleted
		false, // exclusive
		false, // no-wait
		nil,
	); err != nil {
		return fmt.Errorf("declare queue %s failed: %w", name, err)
	}
	c.declaredQueues[name] = struct{}{}
	return nil
}

// DeclareTopicExchange 声明 topic 交换机（幂等，重复声明会被缓存跳过）
func (c *Client) DeclareTopicExchange(name string) error {
	c.declaredMu.Lock()
	defer c.declaredMu.Unlock()

	if _, ok := c.declaredExchanges[name]; ok {
		return nil
	}
	if err := c.channel.ExchangeDeclare(
		name,
		"topic",
		true,  // durable
		false, // auto-deleted
		false, // internal
		false, // no-wait
		nil,
	); err != nil {
		return fmt.Errorf("declare exchange %s failed: %w", name, err)
	}
	c.declaredExchanges[name] = struct{}{}
	return nil
}

// BindQueue 绑定队列到交换机（按 routingKey）
func (c *Client) BindQueue(queue, routingKey, exchange string) error {
	if err := c.DeclareQueue(queue); err != nil {
		return err
	}
	if err := c.DeclareTopicExchange(exchange); err != nil {
		return err
	}
	return c.channel.QueueBind(queue, routingKey, exchange, false, nil)
}

// SimplePublish 使用默认交换机直接路由到指定队列（点对点）
// body 为 JSON 序列化后的字节
func (c *Client) SimplePublish(ctx context.Context, queue string, body []byte) error {
	if err := c.DeclareQueue(queue); err != nil {
		return err
	}
	return c.channel.PublishWithContext(
		ctx,
		"",    // 默认交换机
		queue, // routing key = 队列名
		false, // mandatory
		false, // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Timestamp:    time.Now(),
			Body:         body,
		},
	)
}

// Publish 发布消息到 topic 交换机（按 routingKey 路由）
func (c *Client) Publish(ctx context.Context, exchange, routingKey string, body []byte) error {
	if err := c.DeclareTopicExchange(exchange); err != nil {
		return err
	}
	return c.channel.PublishWithContext(
		ctx,
		exchange,
		routingKey,
		false, // mandatory
		false, // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Timestamp:    time.Now(),
			Body:         body,
		},
	)
}

// Handler 消息处理函数：返回 error 时消息会被 Nack 并重新入队
type Handler func(body []byte) error

// Consume 启动队列消费者（goroutine 内执行 handler，手动 ack）
// 调用方应在插件 Init() 中调用，Close() 时会随连接关闭自动停止
func (c *Client) Consume(ctx context.Context, queue string, handler Handler) error {
	if err := c.DeclareQueue(queue); err != nil {
		return err
	}
	msgs, err := c.channel.ConsumeWithContext(
		ctx,
		queue,
		"",    // consumer tag（自动生成）
		false, // auto-ack：手动 ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,
	)
	if err != nil {
		return fmt.Errorf("consume queue %s failed: %w", queue, err)
	}

	go func() {
		for d := range msgs {
			if err := handler(d.Body); err != nil {
				logger.Error("消息处理失败，重新入队",
					zap.String("queue", queue),
					zap.Error(err))
				// multiple=false, requeue=true：失败重新入队
				_ = d.Nack(false, true)
				continue
			}
			_ = d.Ack(false)
		}
	}()

	logger.Info("RabbitMQ 消费者已启动", zap.String("queue", queue))
	return nil
}

// ErrNotAvailable RabbitMQ 未初始化或不可用
var ErrNotAvailable = errors.New("rabbitmq not available")
