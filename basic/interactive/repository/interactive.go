package repository

import (
	"compus_blog/basic/interactive/domain"
	"compus_blog/basic/interactive/repository/cache"
	"compus_blog/basic/interactive/repository/dao"
	"compus_blog/basic/pkg/logger"
	"context"
	"errors"
	"github.com/ecodeclub/ekit/slice"
)

type InteractiveRepository interface {
	IncrReadCnt(ctx context.Context, biz string, bizId int64) error
	// BatchIncrReadCnt 这里调用者要保证 bizs 和 bizIds 长度一样
	BatchIncrReadCnt(ctx context.Context, bizs []string, bizIds []int64) error
	IncrLike(ctx context.Context, biz string, bizId, uid int64) error
	DecrLike(ctx context.Context, biz string, bizId, uid int64) error
	AddCollectionItem(ctx context.Context, biz string, bizId int64, uid int64, cid int64) error
	Get(ctx context.Context, biz string, bizId int64) (domain.Interactive, error)
	Liked(ctx context.Context, biz string, id int64, uid int64) (bool, error)
	Collected(ctx context.Context, biz string, id int64, uid int64) (bool, error)
	GetByIds(ctx context.Context, biz string, bizIds []int64) ([]domain.Interactive, error)
}

type InteractiveRepositoryStruct struct {
	dao   dao.InteractiveDAO
	cache cache.InteractiveCache
	l     logger.LoggerV1
}

func NewInteractiveRepository(dao dao.InteractiveDAO,
	cache cache.InteractiveCache, l logger.LoggerV1) InteractiveRepository {
	return &InteractiveRepositoryStruct{
		dao:   dao,
		cache: cache,
		l:     l,
	}
}

func (repo *InteractiveRepositoryStruct) IncrReadCnt(ctx context.Context, biz string, bizId int64) error {

	err := repo.dao.IncrReadCnt(ctx, biz, bizId)
	if err != nil {
		return nil
	}
	//可能有数据不一致现象，但是阅读数可以容忍
	return repo.cache.IncrReadCntIfPresent(ctx, biz, bizId)
}

func (repo *InteractiveRepositoryStruct) BatchIncrReadCnt(ctx context.Context, bizs []string, bizIds []int64) error {
	return repo.dao.BatchIncrReadCnt(ctx, bizs, bizIds)
}

func (repo *InteractiveRepositoryStruct) IncrLike(ctx context.Context, biz string, bizId, uid int64) error {
	err := repo.dao.InsertLikeInfo(ctx, biz, bizId, uid)
	if err != nil {
		return err
	}
	return repo.cache.IncrLikeCntIfPresent(ctx, biz, uid)
}

func (repo *InteractiveRepositoryStruct) DecrLike(ctx context.Context, biz string, bizId, uid int64) error {
	err := repo.dao.DeleteLikeInfo(ctx, biz, bizId, uid)
	if err != nil {
		return err
	}
	return repo.cache.DecrLikeCntIfPresent(ctx, biz, uid)
}

//
//func (repo *InteractiveRepositoryStruct) DecrCollectV1(ctx context.Context, biz string, bizId int64, cid int64, uid int64) error {
//	// 判断是否为有效取消收藏
//	ok, err := repo.dao.DeleteCollectInfoV1(ctx, biz, bizId, cid, uid)
//	if err != nil {
//		return err
//	}
//	if ok {
//		return repo.cache.DecrCollectCntIfPresent(ctx, biz, uid)
//	}
//	return nil
//
//}

func (repo *InteractiveRepositoryStruct) AddCollectionItem(ctx context.Context,
	biz string, bizId int64, uid int64, cid int64) error {
	err := repo.dao.InsertCollectionBiz(ctx, dao.UserCollectionBiz{
		Biz:   biz,
		BizId: bizId,
		Uid:   uid,
		Cid:   cid,
	})
	if err != nil {
		return err
	}
	return repo.cache.IncrCollectCntIfPresent(ctx, biz, uid)
}

func (repo *InteractiveRepositoryStruct) Get(ctx context.Context, biz string, bizId int64) (domain.Interactive, error) {
	intr, err := repo.dao.Get(ctx, biz, bizId)
	return repo.toDomain(intr), err
}

func (repo *InteractiveRepositoryStruct) Liked(ctx context.Context, biz string, id int64, uid int64) (bool, error) {
	_, err := repo.dao.GetLikeInfo(ctx, biz, id, uid)
	switch {
	case err == nil:
		return true, nil
	case errors.Is(err, dao.ErrDataNotFound):
		return false, nil
	default:
		return false, err
	}
}

func (repo *InteractiveRepositoryStruct) Collected(ctx context.Context, biz string, id int64, uid int64) (bool, error) {
	_, err := repo.dao.GetCollectInfo(ctx, biz, id, uid)
	switch {
	case err == nil:
		return true, nil
	case errors.Is(err, dao.ErrDataNotFound):
		return false, nil
	default:
		return false, err
	}
}

func (repo *InteractiveRepositoryStruct) toDomain(intr dao.Interactive) domain.Interactive {
	return domain.Interactive{
		Biz:        intr.Biz,
		BizId:      intr.BizId,
		LikeCnt:    intr.LikeCnt,
		CollectCnt: intr.CollectCnt,
		ReadCnt:    intr.ReadCnt,
	}
}

func (repo *InteractiveRepositoryStruct) GetByIds(ctx context.Context, biz string, ids []int64) ([]domain.Interactive, error) {
	intrs, err := repo.dao.GetByIds(ctx, biz, ids)
	if err != nil {
		return nil, err
	}
	return slice.Map[dao.Interactive, domain.Interactive](intrs, func(idx int, src dao.Interactive) domain.Interactive {
		return repo.toDomain(src)
	}), nil
}
