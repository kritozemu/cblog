package article

import (
	logger2 "compus_blog/basic/pkg/logger"
	"compus_blog/basic/pkg/saramax"
	"context"
	"github.com/IBM/sarama"
)

type HistoryReadEventConsumer struct {
	client sarama.Client
	l      logger2.LoggerV1
}

func NewHistoryReadEventConsumer(
	client sarama.Client,
	l logger2.LoggerV1) *HistoryReadEventConsumer {
	return &HistoryReadEventConsumer{
		client: client,
		l:      l,
	}
}

func (r *HistoryReadEventConsumer) Start() error {
	cg, err := sarama.NewConsumerGroupFromClient("history_record",
		r.client)
	if err != nil {
		return err
	}
	go func() {
		err := cg.Consume(context.Background(),
			[]string{"read_article"},
			saramax.NewHandler[ReadEvent](r.l, r.Consume))
		if err != nil {
			r.l.Error("退出了消费循环异常", logger2.Error(err))
		}
	}()
	return err
}

// Consume 这个不是幂等的
func (r *HistoryReadEventConsumer) Consume(msg *sarama.ConsumerMessage, t ReadEvent) error {
	//ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	//defer cancel()
	//return r.repo.Add(ctx, t.Aid, t.Uid)
	panic("implement me")
}
