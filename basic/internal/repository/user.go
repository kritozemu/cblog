package repository

import (
	"compus_blog/basic/internal/domain"
	"compus_blog/basic/internal/repository/cache"
	"compus_blog/basic/internal/repository/dao"
	logger2 "compus_blog/basic/pkg/logger"
	"context"
	"database/sql"
	"errors"
	"time"
)

var (
	ErrDuplicateUser = dao.ErrDuplicateEmail
	ErrUserNotFound  = dao.ErrRecordNotFound
)

type UserRepository interface {
	Create(ctx context.Context, u domain.User) error
	FindByEmail(ctx context.Context, email string) (domain.User, error)
	UpdateNonZeroFields(ctx context.Context, user domain.User) error
	FindByPhone(ctx context.Context, phone string) (domain.User, error)
	FindById(ctx context.Context, uid int64) (domain.User, error)
}

type UserRepositoryStruct struct {
	dao   dao.UserDAO
	cache cache.UserCache
	l     logger2.LoggerV1
}

func NewUserRepository(dao dao.UserDAO, cache cache.UserCache,
	l logger2.LoggerV1) UserRepository {
	return &UserRepositoryStruct{
		dao:   dao,
		cache: cache,
		l:     l,
	}
}

func (r *UserRepositoryStruct) toEntity(u domain.User) dao.User {
	return dao.User{
		Id: u.Id,
		Email: sql.NullString{
			String: u.Email,
			Valid:  u.Email != "",
		},
		Phone: sql.NullString{
			String: u.Phone,
			Valid:  u.Phone != "",
		},
		Password: u.Password,
		Birthday: u.Birthday.UnixMilli(),
		AboutMe:  u.AboutMe,
		Nickname: u.Nickname,
	}
}

func (r *UserRepositoryStruct) toDomain(u dao.User) domain.User {
	return domain.User{
		Id:       u.Id,
		Email:    u.Email.String,
		Phone:    u.Phone.String,
		Password: u.Password,
		Birthday: time.UnixMilli(u.Birthday),
		AboutMe:  u.AboutMe,
		Nickname: u.Nickname,
		Ctime:    time.UnixMilli(u.Ctime),
	}
}

func (r *UserRepositoryStruct) Create(ctx context.Context, u domain.User) error {
	return r.dao.Insert(ctx, r.toEntity(u))
}

func (r *UserRepositoryStruct) FindByEmail(ctx context.Context, email string) (domain.User, error) {
	u, err := r.dao.FindByEmail(ctx, email)
	if err != nil {
		return domain.User{}, err
	}
	return r.toDomain(u), nil
}

func (r *UserRepositoryStruct) UpdateNonZeroFields(ctx context.Context, user domain.User) error {
	err := r.dao.UpdateById(ctx, r.toEntity(user))
	if err != nil {
		return err
	}
	// 延迟一秒
	time.AfterFunc(time.Second, func() {
		_ = r.cache.Del(ctx, user.Id)
	})
	return r.cache.Del(ctx, user.Id)
}

func (r *UserRepositoryStruct) FindByPhone(ctx context.Context, phone string) (domain.User, error) {
	user, err := r.dao.FindByPhone(ctx, phone)
	if err != nil {
		return domain.User{}, err
	}
	return r.toDomain(user), nil
}

func (r *UserRepositoryStruct) FindById(ctx context.Context, uid int64) (domain.User, error) {
	du, err := r.cache.Get(ctx, uid)
	if err == nil {
		return du, err
	}

	// 检测限流/熔断/降级标记位
	if ctx.Value("downgrade") == "true" {
		return du, errors.New("触发降级，不再查询数据库")
	}

	u, err := r.dao.FindById(ctx, uid)
	if err != nil {
		return domain.User{}, err
	}

	err = r.cache.Set(ctx, du)
	if err != nil {
		// 网络崩了，也可能是 redis 崩了
		r.l.Error("网络或者redis崩溃", logger2.Error(err))
	}

	return r.toDomain(u), nil
}
