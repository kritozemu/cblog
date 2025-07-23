package dao

import (
	"context"
	"errors"
	"time"
)

var ErrPossibleIncorrectAuthor = errors.New("用户在尝试操作非本人数据")

type ArticleDAO interface {
	Insert(ctx context.Context, art Article) (int64, error)
	UpdateById(ctx context.Context, art Article) error
	SyncStatus(ctx context.Context, id int64, author int64, status uint8) error
	Sync(ctx context.Context, art Article) (int64, error)
	GetByAuthor(ctx context.Context, uid int64, offset int, limit int) ([]Article, error)
	ListPub(ctx context.Context, start time.Time, offset int, limit int) ([]Article, error)
	GetById(ctx context.Context, id int64) (Article, error)
	GetPublishedById(ctx context.Context, id int64) (PublishedArticle, error)
}
