package service

import (
	"compus_blog/basic/internal/domain"
	"compus_blog/basic/internal/repository"
	"context"
	"errors"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrDuplicateEmail        = repository.ErrDuplicateUser
	ErrInvalidUserOrPassword = errors.New("用户不存在或者密码不对")
)

type UserService interface {
	Signup(ctx context.Context, u domain.User) error
	Login(ctx context.Context, email string, password string) (domain.User, error)
	UpdateNonSensitiveInfo(ctx context.Context,
		user domain.User) error
	FindById(ctx context.Context,
		uid int64) (domain.User, error)
	FindOrCreate(ctx context.Context, phone string) (domain.User, error)
}

type UserServiceStruct struct {
	repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) UserService {
	return &UserServiceStruct{
		repo: repo,
	}
}

func (s *UserServiceStruct) Signup(ctx context.Context, u domain.User) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hash)
	return s.repo.Create(ctx, u)
}

func (s *UserServiceStruct) Login(ctx context.Context, email string, password string) (domain.User, error) {
	user, err := s.repo.FindByEmail(ctx, email)
	if errors.Is(err, repository.ErrUserNotFound) {
		return domain.User{}, ErrInvalidUserOrPassword
	}
	if err != nil {
		return domain.User{}, err
	}
	// 检查密码对不对
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return domain.User{}, ErrInvalidUserOrPassword
	}
	return user, nil
}

func (s *UserServiceStruct) UpdateNonSensitiveInfo(ctx context.Context, user domain.User) error {
	// UpdateNicknameAndXXAnd
	return s.repo.UpdateNonZeroFields(ctx, user)
}

func (s *UserServiceStruct) FindById(ctx context.Context, uid int64) (domain.User, error) {
	return s.repo.FindById(ctx, uid)
}

func (s *UserServiceStruct) FindOrCreate(ctx context.Context, phone string) (domain.User, error) {
	// 先尝试找一下用户
	user, err := s.repo.FindByPhone(ctx, phone)
	if !errors.Is(err, repository.ErrUserNotFound) {
		// 有两种情况
		// err == nil, u 是可用的
		// err != nil，系统错误，
		return user, err
	}
	//用户没找到
	err = s.repo.Create(ctx, domain.User{
		Phone: phone,
	})
	//系统错误或者是索引冲突
	if err != nil && !errors.Is(err, repository.ErrUserNotFound) {
		return domain.User{}, err
	}

	return s.repo.FindByPhone(ctx, phone)
}
