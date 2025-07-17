package cache

import (
	"compus_blog/basic/internal/domain"
	"context"
	"encoding/json"
	"fmt"
	"github.com/redis/go-redis/v9"
	"time"
)

var ErrKeyNotExist = redis.Nil

type UserCache interface {
	Get(ctx context.Context, uid int64) (domain.User, error)
	Set(ctx context.Context, du domain.User) error
	Del(ctx context.Context, id int64) error
}

type RedisCachedUser struct {
	cmd        redis.Cmdable
	expiration time.Duration
}

func NewUserCache(cmd redis.Cmdable) UserCache {
	return &RedisCachedUser{
		cmd:        cmd,
		expiration: time.Minute * 15,
	}
}

func (r *RedisCachedUser) Get(ctx context.Context, uid int64) (domain.User, error) {
	key := r.key(uid)
	data, err := r.cmd.Get(ctx, key).Bytes()
	if err != nil {
		return domain.User{}, err
	}
	var user domain.User
	err = json.Unmarshal(data, &user)
	return user, err
}

func (r *RedisCachedUser) Set(ctx context.Context, du domain.User) error {
	key := r.key(du.Id)
	data, err := json.Marshal(du)
	if err != nil {
		return err
	}
	return r.cmd.Set(ctx, key, data, r.expiration).Err()
}

func (r *RedisCachedUser) Del(ctx context.Context, id int64) error {
	key := r.key(id)
	return r.cmd.Del(ctx, key).Err()
}

func (r *RedisCachedUser) key(uid int64) string {
	return fmt.Sprintf("user:%d", uid)
}
