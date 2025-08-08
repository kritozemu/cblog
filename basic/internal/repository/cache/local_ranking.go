package cache

import (
	"compus_blog/basic/internal/domain"
	"context"
	"errors"
	"github.com/ecodeclub/ekit/syncx/atomicx"
	"time"
)

type RankingLocalCache struct {
	topN     *atomicx.Value[[]domain.Article]
	ddl      *atomicx.Value[time.Time]
	duration time.Duration
}

func NewRankingLocalCache() *RankingLocalCache {
	return &RankingLocalCache{
		topN:     atomicx.NewValue[[]domain.Article](),
		ddl:      atomicx.NewValueOf[time.Time](time.Now()),
		duration: time.Minute * 10,
	}
}

func (c *RankingLocalCache) Get(ctx context.Context) ([]domain.Article, error) {
	ddl := c.ddl.Load()
	arts := c.topN.Load()
	if len(arts) == 0 || ddl.Before(time.Now()) {
		return nil, errors.New("缓存未命中")
	}
	return arts, nil
}

func (c *RankingLocalCache) Set(ctx context.Context, data []domain.Article) error {
	c.topN.Store(data)
	ddl := time.Now().Add(c.duration)
	c.ddl.Store(ddl)
	return nil
}

func (c *RankingLocalCache) ForceGet(ctx context.Context) ([]domain.Article, error) {
	arts := c.topN.Load()
	return arts, nil
}
