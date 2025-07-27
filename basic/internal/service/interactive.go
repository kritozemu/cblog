package service

import (
	"compus_blog/basic/internal/domain"
	"compus_blog/basic/internal/pkg/logger"
	"compus_blog/basic/internal/repository"
	"context"
	"golang.org/x/sync/errgroup"
)

type InteractiveService interface {
	Get(ctx context.Context, biz string, bizId int64, uid int64) (domain.Interactive, error)
	IncrReadCnt(ctx context.Context, biz string, bizId int64) error
	Like(ctx context.Context, biz string, id int64, uid int64) error
	CancelLike(ctx context.Context, biz string, id int64, uid int64) error
	Collect(ctx context.Context, biz string, id, cid, uid int64) error
	//CancelCollect(ctx context.Context, biz string, id int64, uid int64) error
	//CancelCollectV1(ctx context.Context, biz string, id int64, cid int64, uid int64) error
	GetByIds(ctx context.Context, biz string, bizIds []int64) (map[int64]domain.Interactive, error)
}

type InteractiveServiceStruct struct {
	repo repository.InteractiveRepository
	l    logger.LoggerV1
}

func NewInteractiveService(repo repository.InteractiveRepository,
	l logger.LoggerV1) InteractiveService {
	return &InteractiveServiceStruct{
		repo: repo,
		l:    l,
	}
}

func (svc *InteractiveServiceStruct) Get(ctx context.Context, biz string, bizId int64, uid int64) (domain.Interactive, error) {
	intr, err := svc.repo.Get(ctx, biz, bizId)
	if err != nil {
		return domain.Interactive{}, err
	}
	var eg errgroup.Group
	eg.Go(func() error {
		intr.Liked, err = svc.repo.Liked(ctx, biz, bizId, uid)
		return err
	})

	eg.Go(func() error {
		intr.Collected, err = svc.repo.Collected(ctx, biz, bizId, uid)
		return err
	})

	err = eg.Wait()
	if err != nil {
		// 这个查询失败只需要记录日志就可以，不需要中断执行
		svc.l.Error("查询用户是否点赞的信息失败",
			logger.String("biz", biz),
			logger.Int64("bizId", bizId),
			logger.Int64("uid", uid),
			logger.Error(err))
	}
	return intr, nil
}

func (svc *InteractiveServiceStruct) IncrReadCnt(ctx context.Context, biz string, bizId int64) error {
	return svc.repo.IncrReadCnt(ctx, biz, bizId)
}

func (svc *InteractiveServiceStruct) Like(ctx context.Context, biz string, id int64, uid int64) error {
	return svc.repo.IncrLike(ctx, biz, id, uid)
}

func (svc *InteractiveServiceStruct) CancelLike(ctx context.Context, biz string, id int64, uid int64) error {
	return svc.repo.DecrLike(ctx, biz, id, uid)
}

func (svc *InteractiveServiceStruct) Collect(ctx context.Context, biz string, id, cid, uid int64) error {
	return svc.repo.AddCollectionItem(ctx, biz, id, cid, uid)
}

func (svc *InteractiveServiceStruct) CancelCollect(ctx context.Context, biz string, id int64, uid int64) error {
	return svc.repo.DecrCollect(ctx, biz, id, uid)
}

//
//func (svc *InteractiveServiceStruct) CancelCollectV1(ctx context.Context, biz string, id int64, cid int64, uid int64) error {
//	return svc.repo.DecrCollectV1(ctx, biz, id, cid, uid)
//}

func (svc *InteractiveServiceStruct) GetByIds(ctx context.Context, biz string, bizIds []int64) (map[int64]domain.Interactive, error) {
	intrs, err := svc.repo.GetByIds(ctx, biz, bizIds)
	if err != nil {
		return nil, err
	}
	res := make(map[int64]domain.Interactive, len(intrs))
	for _, intr := range intrs {
		res[intr.BizId] = intr
	}
	return res, nil
}
