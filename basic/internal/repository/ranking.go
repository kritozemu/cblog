package repository

import (
	"compus_blog/basic/internal/domain"
	"compus_blog/basic/internal/repository/cache"
	"context"
)

type RankingRepository interface {
	ReplaceTopN(ctx context.Context, arts []domain.Article) error
	GetTopN(ctx context.Context) ([]domain.Article, error)
}

type RankingRepositoryStruct struct {
	redis *cache.RankingRedisCache
	local *cache.RankingLocalCache
}

func NewRankingRepository(redis *cache.RankingRedisCache, local *cache.RankingLocalCache) RankingRepository {
	return &RankingRepositoryStruct{
		redis: redis,
		local: local,
	}
}

func (r *RankingRepositoryStruct) GetTopN(ctx context.Context) ([]domain.Article, error) {
	data, err := r.local.Get(ctx)
	if err == nil {
		return nil, err
	}
	data, err = r.redis.Get(ctx)
	if err == nil {
		r.local.Set(ctx, data)
	} else {
		return r.local.ForceGet(ctx)
	}

	return data, err

}

func (r *RankingRepositoryStruct) ReplaceTopN(ctx context.Context, arts []domain.Article) error {
	_ = r.local.Set(ctx, arts)
	return r.redis.Set(ctx, arts)
}
