package article

import (
	logger2 "compus_blog/basic/pkg/logger"
	"context"
	"encoding/json"
	"github.com/IBM/sarama"
	"math/rand"
	"time"
)

type Producer interface {
	ProduceReadEvent(ctx context.Context, evt ReadEvent) error
	BatchProduceReadEventV1(ctx context.Context, evts []ReadEvent)
}

type KafkaProducer struct {
	producer sarama.SyncProducer
	l        logger2.LoggerV1
}

func NewKafkaProducer(producer sarama.SyncProducer, l logger2.LoggerV1) Producer {
	return &KafkaProducer{
		producer: producer,
		l:        l,
	}
}

func (p *KafkaProducer) ProduceReadEvent(ctx context.Context, evt ReadEvent) error {
	data, err := json.Marshal(evt)
	if err != nil {
		return err
	}
	_, _, err = p.producer.SendMessage(&sarama.ProducerMessage{
		Topic: "read_article",
		Value: sarama.ByteEncoder(data),
	})
	return err
}

func (p *KafkaProducer) BatchProduceReadEventV1(ctx context.Context, evts []ReadEvent) {
	msgs := make([]*sarama.ProducerMessage, 0, len(evts))
	failedEvts := make([]ReadEvent, 0) // 记录序列化失败的事件

	for _, evt := range evts {
		data, err := json.Marshal(evt)
		if err != nil {
			p.l.Error("消息序列化失败", logger2.Error(err),
				logger2.Int64("uid:", evt.Uid), logger2.Int64("aid:", evt.Aid))
			failedEvts = append(failedEvts, evt) // 收集失败事件
			continue                             // 直接跳过，不加入批次
		}
		msgs = append(msgs, &sarama.ProducerMessage{
			Topic: "read_article",
			Value: sarama.ByteEncoder(data),
		})
	}

	// 处理失败事件（如持久化到本地，后续人工介入）
	if len(failedEvts) > 0 {
		p.saveFailedEvents(failedEvts) // 自定义方法：保存到文件或数据库
	}

	// 仅发送有效消息
	if len(msgs) == 0 {
		return
	}

	// 批量发送重试（3次，带退避）
	var lastErr error
	const maxRetry = 3
	for i := 0; i < maxRetry; i++ {
		lastErr = p.producer.SendMessages(msgs)
		if lastErr == nil {
			return
		}
		// 若上下文已取消，停止重试
		if ctx.Err() != nil {
			p.l.Error("上下文已取消,停止重试kafka的消息发送", logger2.Error(ctx.Err()))
			return
		}
		// 退避延迟（指数退避+随机抖动）
		delay := time.Duration(200*i) * time.Millisecond
		delay += time.Duration(rand.Intn(100)) * time.Millisecond // 避免重试时间重叠
		select {
		case <-time.After(delay):
		case <-ctx.Done():
			return
		}
	}

	// 多次重试失败后，将整个批次消息加入失败存储
	p.saveFailedBatch(msgs)
	p.l.Error("批量消息发送失败", logger2.Int64("已重试", maxRetry))
}

// 通过日志告警、持久化到本地文件 / 数据库等方式兜底
func (p *KafkaProducer) saveFailedEvents(evts []ReadEvent) {
	//todo
	panic("implement me")
}

// 需实现：将失败的批量消息持久化
func (p *KafkaProducer) saveFailedBatch(msgs []*sarama.ProducerMessage) {
	//todo
	panic("implement me")
}

type ReadEvent struct {
	Uid int64
	Aid int64
}
