package service

import (
	intrv1 "compus_blog/basic/api/proto/gen/intr/v1"
	"compus_blog/basic/internal/domain"
	"compus_blog/basic/internal/repository"
	"context"
	"errors"
	"github.com/ecodeclub/ekit/queue"
	"github.com/ecodeclub/ekit/slice"
	"math"
	"time"
)

type RankingService interface {
	TopN(ctx context.Context) error
}

type BatchRankingService struct {
	artSvc  ArticleService
	intrSvc intrv1.InteractiveServiceClient
	repo    repository.RankingRepository

	batchSize int
	n         int
	//算法
	scoreFunc func(t time.Time, likeCnt int64) float64

	//负载
	load int64
}

func NewBatchRankingService(artSvc ArticleService, intrSvc intrv1.InteractiveServiceClient,
	repo repository.RankingRepository) RankingService {
	return &BatchRankingService{
		artSvc:    artSvc,
		intrSvc:   intrSvc,
		repo:      repo,
		batchSize: 100,
		n:         100,
		scoreFunc: func(t time.Time, likeCnt int64) float64 {
			sec := time.Since(t).Seconds()
			return float64(likeCnt-1) / math.Pow(sec+2.0, 1.5)
		},
	}
}

func (svc *BatchRankingService) TopN(ctx context.Context) error {
	arts, err := svc.topN(ctx)
	if err != nil {
		return err
	}
	return svc.repo.ReplaceTopN(ctx, arts)
}

func (svc *BatchRankingService) topN(ctx context.Context) ([]domain.Article, error) {
	now := time.Now()
	// 先拿一批数据
	offset := 0
	type Score struct {
		art   domain.Article
		score float64
	}
	// 这里可以用非并发安全
	topN := queue.NewConcurrentPriorityQueue[Score](svc.n,
		func(src Score, dst Score) int {
			if src.score > dst.score {
				return 1
			} else if src.score == dst.score {
				return 0
			} else {
				return -1
			}
		})

	for {
		//这里先拿一批
		arts, err := svc.artSvc.ListPub(ctx, now, offset, svc.batchSize)
		if err != nil {
			return nil, err
		}

		ids := slice.Map[domain.Article, int64](arts, func(idx int, src domain.Article) int64 {
			return src.Id
		})

		intrs, err := svc.intrSvc.GetByIds(ctx, &intrv1.GetByIdsRequest{
			Biz: "article", Ids: ids,
		})
		if err != nil {
			return nil, err
		}
		if len(intrs.Intrs) == 0 {
			return nil, errors.New("没有数据")
		}
		// 合并计算 score
		// 排序
		for _, art := range arts {
			intr, ok := intrs.Intrs[art.Id]
			if !ok {
				// 你都没有，肯定不可能是热榜
				continue
			}

			score := svc.scoreFunc(art.Utime, intr.LikeCnt)
			err = topN.Enqueue(Score{
				art:   art,
				score: score,
			})
			if errors.Is(err, queue.ErrOutOfCapacity) {
				val, _ := topN.Dequeue()
				if val.score < score {
					_ = topN.Enqueue(Score{
						art:   art,
						score: score,
					})
				} else {
					_ = topN.Enqueue(val)
				}
			}
		}
		// 一批已经处理完了
		if len(arts) < svc.batchSize ||
			now.Sub(arts[len(arts)-1].Utime).Hours() > 7*24 {
			// 我这一批都没取够，我当然可以肯定没有下一批了
			// 又或者已经取到了七天之前的数据了，说明可以中断了
			break
		}
		// 依旧更新offset
		offset += len(arts)

	}
	res := make([]domain.Article, svc.n)
	for i := 0; i < svc.n; i++ {
		val, err := topN.Dequeue()
		if err != nil {
			break
		}
		res[i] = val.art
	}
	return res, nil
}
