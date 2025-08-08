////go:build need_fix

package service

import (
	domain2 "compus_blog/basic/interactive/domain"
	"compus_blog/basic/interactive/service"
	svcmocks2 "compus_blog/basic/interactive/service/mocks"
	"compus_blog/basic/internal/domain"
	"compus_blog/basic/internal/repository"
	repomocks "compus_blog/basic/internal/repository/mocks"
	svcmocks "compus_blog/basic/internal/service/mocks"
	"context"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"testing"
	"time"
)

func TestRankingTopN(t *testing.T) {
	now := time.Now()
	testCases := []struct {
		name string
		mock func(ctrl *gomock.Controller) (ArticleService,
			service.InteractiveService,
			repository.RankingRepository)

		wantErr  error
		wantArts []domain.Article
	}{
		{
			name: "计算成功",
			// 怎么模拟我的数据？
			mock: func(ctrl *gomock.Controller) (ArticleService, service.InteractiveService,
				repository.RankingRepository) {
				artSvc := svcmocks.NewMockArticleService(ctrl)
				// 最简单，一批就搞完
				artSvc.EXPECT().ListPub(gomock.Any(), gomock.Any(), 0, 3).
					Return([]domain.Article{
						{Id: 1, Utime: now, Ctime: now},
						{Id: 2, Utime: now, Ctime: now},
						{Id: 3, Utime: now, Ctime: now},
					}, nil)
				artSvc.EXPECT().ListPub(gomock.Any(), gomock.Any(), 3, 3).
					Return([]domain.Article{}, nil)

				intrSvc := svcmocks2.NewMockInteractiveService(ctrl)
				intrSvc.EXPECT().GetByIds(gomock.Any(),
					"article", []int64{1, 2, 3}).
					Return(map[int64]domain2.Interactive{
						1: {BizId: 1, LikeCnt: 1},
						2: {BizId: 2, LikeCnt: 2},
						3: {BizId: 3, LikeCnt: 3},
					}, nil)
				intrSvc.EXPECT().GetByIds(gomock.Any(),
					"article", []int64{}).
					Return(map[int64]domain2.Interactive{}, nil)

				rankRepo := repomocks.NewMockRankingRepository(ctrl)
				return artSvc, intrSvc, rankRepo
			},
			wantArts: []domain.Article{
				{Id: 3, Utime: now, Ctime: now},
				{Id: 2, Utime: now, Ctime: now},
				{Id: 1, Utime: now, Ctime: now},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			artSvc, intrSvc, rankRepo := tc.mock(ctrl)
			svc := NewBatchRankingServiceV1(artSvc, intrSvc, rankRepo).(*BatchRankingServiceV1)
			// 为了测试
			svc.batchSize = 3
			svc.n = 3
			svc.scoreFunc = func(t time.Time, likeCnt int64) float64 {
				return float64(likeCnt)
			}
			arts, err := svc.topNV1(context.Background())
			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.wantArts, arts)
		})
	}
}
