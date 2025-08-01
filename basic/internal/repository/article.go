package repository

import (
	"compus_blog/basic/internal/domain"
	"compus_blog/basic/internal/repository/cache"
	"compus_blog/basic/internal/repository/dao"
	logger2 "compus_blog/basic/pkg/logger"
	"context"
	"github.com/ecodeclub/ekit/slice"
	"time"
)

type ArticleRepository interface {
	Create(ctx context.Context, art domain.Article) (int64, error)
	Update(ctx context.Context, art domain.Article) error
	SyncStatus(ctx context.Context, art domain.Article) error
	Sync(ctx context.Context, art domain.Article) (int64, error)
	List(ctx context.Context, uid int64, offset int, limit int) ([]domain.Article, error)
	ListPub(ctx context.Context, start time.Time, offset int, limit int) ([]domain.Article, error)
	GetById(ctx context.Context, id int64) (domain.Article, error)
	GetPublishedById(ctx context.Context, id int64) (domain.Article, error)
}

type ArticleRepositoryStruct struct {
	dao     dao.ArticleDAO
	cache   cache.ArticleCache
	usrRepo UserRepository
	l       logger2.LoggerV1
}

func NewArticleRepository(dao dao.ArticleDAO, cache cache.ArticleCache,
	usrRepo UserRepository, l logger2.LoggerV1) ArticleRepository {
	return &ArticleRepositoryStruct{
		dao:     dao,
		cache:   cache,
		usrRepo: usrRepo,
		l:       l,
	}
}

func (a *ArticleRepositoryStruct) Create(ctx context.Context, art domain.Article) (int64, error) {
	defer func() {
		a.cache.DelFirstPage(ctx, art.Author.Id)
	}()
	return a.dao.Insert(ctx, dao.Article{
		Title:    art.Title,
		Content:  art.Content,
		AuthorId: art.Author.Id,
		Status:   uint8(art.Status),
	})
}

func (a *ArticleRepositoryStruct) Update(ctx context.Context, art domain.Article) error {
	defer func() {
		a.cache.DelFirstPage(ctx, art.Author.Id)
	}()
	return a.dao.UpdateById(ctx, dao.Article{
		Id:       art.Id,
		Title:    art.Title,
		Content:  art.Content,
		AuthorId: art.Author.Id,
		Status:   uint8(art.Status),
	})
}

func (a *ArticleRepositoryStruct) SyncStatus(ctx context.Context, art domain.Article) error {
	return a.dao.SyncStatus(ctx, art.Id, art.Author.Id, uint8(art.Status))
}

func (a *ArticleRepositoryStruct) Sync(ctx context.Context, art domain.Article) (int64, error) {
	id, err := a.dao.Sync(ctx, a.toEntity(art))
	if err == nil {
		a.cache.DelFirstPage(ctx, art.Author.Id)
		er := a.cache.SetPub(ctx, art)
		if er != nil {
			// 不需要特别关心
			// 比如说输出 WARN 日志
		}
	}

	// 在这里尝试，设置缓存
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		// 你可以灵活设置过期时间
		user, er := a.usrRepo.FindById(ctx, art.Author.Id)
		if er != nil {
			// 要记录日志
			return
		}
		art.Author = domain.Author{
			Id:   user.Id,
			Name: user.Nickname,
		}
		er = a.cache.SetPub(ctx, art)
		if er != nil {
			// 记录日志
		}
	}()
	return id, err
}

func (a *ArticleRepositoryStruct) toEntity(article domain.Article) dao.Article {
	return dao.Article{
		Id:       article.Id,
		Title:    article.Title,
		Content:  article.Content,
		AuthorId: article.Author.Id,
		Status:   uint8(article.Status),
	}
}

func (a *ArticleRepositoryStruct) toDomain(article dao.Article) domain.Article {
	return domain.Article{
		Id:      article.Id,
		Title:   article.Title,
		Content: article.Content,
		Status:  domain.ArticleStatus(article.Status),
		Author: domain.Author{
			Id: article.AuthorId,
		},
		Ctime: time.UnixMilli(article.Ctime),
		Utime: time.UnixMilli(article.Utime),
	}
}

func (a *ArticleRepositoryStruct) List(ctx context.Context, uid int64, offset int, limit int) ([]domain.Article, error) {
	if offset == 0 && limit <= 100 {
		data, err := a.cache.GetFirstPage(ctx, uid)
		if err == nil {
			go func() {
				a.preCache(ctx, data)
			}()
			return data, nil
		}

	}

	res, err := a.dao.GetByAuthor(ctx, uid, offset, limit)
	if err != nil {
		return nil, err
	}

	data := slice.Map[dao.Article, domain.Article](res, func(idx int, src dao.Article) domain.Article {
		return a.toDomain(src)
	})

	//回写缓存
	go func() {
		er := a.cache.SetFirstPage(ctx, uid, data)
		if er != nil {
			a.l.Error("回写缓存失败", logger2.Error(err))
		}
		a.preCache(ctx, data)
	}()
	return data, nil
}

func (a *ArticleRepositoryStruct) ListPub(ctx context.Context, start time.Time, offset int, limit int) ([]domain.Article, error) {
	res, err := a.dao.ListPub(ctx, start, offset, limit)
	if err != nil {
		return nil, err
	}
	return slice.Map[dao.Article, domain.Article](res, func(idx int, src dao.Article) domain.Article {
		return a.toDomain(src)
	}), nil
}

func (a *ArticleRepositoryStruct) GetById(ctx context.Context, id int64) (domain.Article, error) {
	artd, err := a.dao.GetById(ctx, id)
	if err != nil {
		return domain.Article{}, err
	}
	return a.toDomain(artd), nil
}

func (a *ArticleRepositoryStruct) GetPublishedById(ctx context.Context, id int64) (domain.Article, error) {
	art, err := a.dao.GetPublishedById(ctx, id)
	if err != nil {
		return domain.Article{}, err
	}
	usr, err := a.usrRepo.FindById(ctx, art.AuthorId)
	if err != nil {
		a.l.Error("未知的作者", logger2.Error(err))
		return domain.Article{}, err
	}
	res := domain.Article{
		Id:      art.Id,
		Title:   art.Title,
		Content: art.Content,
		Status:  domain.ArticleStatus(art.Status),
		Author: domain.Author{
			Id:   usr.Id,
			Name: usr.Nickname,
		},
		Ctime: time.UnixMilli(art.Ctime),
		Utime: time.UnixMilli(art.Utime),
	}
	return res, nil
}

func (a *ArticleRepositoryStruct) preCache(ctx context.Context, data []domain.Article) {
	if len(data) > 0 && len(data[0].Content) < 1024*1024 {
		err := a.cache.Set(ctx, data[0])
		if err != nil {
			a.l.Error("提前预加载失败", logger2.Error(err))
		}
	}
}
