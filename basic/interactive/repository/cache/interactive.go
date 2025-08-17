package cache

import (
	"compus_blog/basic/interactive/domain"
	"context"
	_ "embed"
	"fmt"
	"github.com/redis/go-redis/v9"
	"strconv"
	"time"
)

var (
	//go:embed lua/interactive_incr_cnt.lua
	luaIncrCnt string
)

const (
	fieldReadCnt    = "read_cnt"
	fieldLikeCnt    = "like_cnt"
	fieldCollectCnt = "collect_cnt"
)

type InteractiveCache interface {
	// IncrReadCntIfPresent 如果在缓存中有对应的数据，就 +1
	IncrReadCntIfPresent(ctx context.Context,
		biz string, bizId int64) error
	IncrLikeCntIfPresent(ctx context.Context,
		biz string, bizId int64) error
	DecrLikeCntIfPresent(ctx context.Context,
		biz string, bizId int64) error
	IncrCollectCntIfPresent(ctx context.Context, biz string, bizId int64) error
	DecrCollectCntIfPresent(ctx context.Context, biz string, bizId int64) error
	// Get 查询缓存中数据
	Get(ctx context.Context, biz string, bizId int64) (domain.Interactive, error)
	Set(ctx context.Context, biz string, bizId int64, intr domain.Interactive) error
}

type InteractiveCacheStruct struct {
	cmd redis.Cmdable
}

func NewInteractiveCache(cmd redis.Cmdable) InteractiveCache {
	return &InteractiveCacheStruct{cmd: cmd}
}

func (r *InteractiveCacheStruct) IncrCollectCntIfPresent(ctx context.Context,
	biz string, bizId int64) error {
	return r.cmd.Eval(ctx, luaIncrCnt,
		[]string{r.key(biz, bizId)},
		fieldCollectCnt, 1).Err()
}

func (r *InteractiveCacheStruct) DecrCollectCntIfPresent(ctx context.Context,
	biz string, bizId int64) error {
	return r.cmd.Eval(ctx, luaIncrCnt,
		[]string{r.key(biz, bizId)},
		fieldCollectCnt, -1).Err()
}

func (r *InteractiveCacheStruct) IncrReadCntIfPresent(ctx context.Context,
	biz string, bizId int64) error {
	return r.cmd.Eval(ctx, luaIncrCnt,
		[]string{r.key(biz, bizId)},
		fieldReadCnt, 1).Err()
}

func (r *InteractiveCacheStruct) IncrLikeCntIfPresent(ctx context.Context,
	biz string, bizId int64) error {
	return r.cmd.Eval(ctx, luaIncrCnt,
		[]string{r.key(biz, bizId)},
		fieldLikeCnt, 1).Err()
}

func (r *InteractiveCacheStruct) DecrLikeCntIfPresent(ctx context.Context,
	biz string, bizId int64) error {
	return r.cmd.Eval(ctx, luaIncrCnt,
		[]string{r.key(biz, bizId)},
		fieldLikeCnt, -1).Err()
}

func (r *InteractiveCacheStruct) Get(ctx context.Context, biz string, bizId int64) (domain.Interactive, error) {
	key := r.key(biz, bizId)
	data, err := r.cmd.HGetAll(ctx, key).Result()
	if err != nil {
		return domain.Interactive{}, err
	}

	//判断是否为空
	if len(data) == 0 {
		return domain.Interactive{}, nil
	}

	var intr = domain.Interactive{
		Biz:   biz,
		BizId: bizId,
	}
	intr.ReadCnt, _ = strconv.ParseInt(data[fieldReadCnt], 10, 64)
	intr.LikeCnt, _ = strconv.ParseInt(data[fieldLikeCnt], 10, 64)
	intr.CollectCnt, _ = strconv.ParseInt(data[fieldCollectCnt], 10, 64)

	return intr, nil

}

func (r *InteractiveCacheStruct) Set(ctx context.Context, biz string, bizId int64, intr domain.Interactive) error {
	key := r.key(biz, bizId)
	err := r.cmd.HMSet(ctx, key,
		fieldReadCnt, intr.ReadCnt,
		fieldLikeCnt, intr.LikeCnt,
		fieldCollectCnt, intr.CollectCnt).Err()
	if err != nil {
		return err
	}
	return r.cmd.Expire(ctx, key, time.Minute*15).Err()
}

func (r *InteractiveCacheStruct) key(biz string, bizId int64) string {
	return fmt.Sprintf("interactive:%s:%d", biz, bizId)
}
