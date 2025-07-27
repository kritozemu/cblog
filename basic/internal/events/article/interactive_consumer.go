package article

import (
	"compus_blog/basic/internal/pkg/logger"
	"compus_blog/basic/internal/pkg/saramax"
	"compus_blog/basic/internal/repository"
	"context"
	"github.com/IBM/sarama"
	"time"
)

type InteractiveReadEventBatchConsumer struct {
	client sarama.Client
	repo   repository.InteractiveRepository
	l      logger.LoggerV1
}

func NewInteractiveReadEventBatchConsumer(client sarama.Client, l logger.LoggerV1,
	repo repository.InteractiveRepository) *InteractiveReadEventBatchConsumer {
	return &InteractiveReadEventBatchConsumer{
		client: client,
		l:      l,
		repo:   repo,
	}
}

func (c *InteractiveReadEventBatchConsumer) Start() error {
	cg, err := sarama.NewConsumerGroupFromClient("interactive", c.client)
	if err != nil {
		return err
	}
	go func() {
		err = cg.Consume(context.Background(), []string{"read_article"},
			saramax.NewBatchHandler[ReadEvent](c.l, c.Consume))
		if err != nil {
			c.l.Error("退出了消费循环异常", logger.Error(err))
		}
	}()
	return err
}

func (c *InteractiveReadEventBatchConsumer) Consume(msg []*sarama.ConsumerMessage, evts []ReadEvent) error {
	ids := make([]int64, 0, len(evts))
	bizs := make([]string, 0, len(evts))
	for _, evt := range evts {
		ids = append(ids, evt.Aid)
		bizs = append(bizs, "article")
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	err := c.repo.BatchIncrReadCnt(ctx, bizs, ids)
	if err != nil {
		c.l.Error("批量增加阅读计数失败",
			logger.Field{Key: "ids", Value: ids},
			logger.Error(err))
	}
	return nil
}
