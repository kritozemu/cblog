package cache

import (
	"compus_blog/basic/internal/domain"
	"context"
	"encoding/json"
	"github.com/redis/go-redis/v9"
	"time"
)

type RankingCache interface {
	Get(ctx context.Context) ([]domain.Article, error)
	Set(ctx context.Context, arts []domain.Article) error
}

type RankingRedisCache struct {
	client redis.Cmdable
	key    string
}

func NewRankingRedisCache(client redis.Cmdable) *RankingRedisCache {
	return &RankingRedisCache{client: client, key: "ranking"}
}

func (c *RankingRedisCache) Get(ctx context.Context) ([]domain.Article, error) {
	data, err := c.client.Get(ctx, c.key).Bytes()
	if err != nil {
		return nil, err
	}
	var res []domain.Article
	err = json.Unmarshal(data, &res)
	return res, err
}

func (c *RankingRedisCache) Set(ctx context.Context, arts []domain.Article) error {
	for i := 0; i < len(arts); i++ {
		arts[i].Content = ""
	}

	data, err := json.Marshal(arts)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, c.key, data, 10*time.Minute).Err()
}
