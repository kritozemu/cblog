package service

import (
	"compus_blog/basic/internal/domain"
	"compus_blog/basic/internal/repository"
	"context"
	"time"
)

type ArticleService interface {
	Save(ctx context.Context, art domain.Article) (int64, error)
	Withdraw(ctx context.Context, art domain.Article) error
	Publish(ctx context.Context, art domain.Article) (int64, error)
	List(ctx context.Context, uid int64, offset int, limit int) ([]domain.Article, error)
	// ListPub 根据这个 start 时间来查询
	ListPub(ctx context.Context, start time.Time, offset, limit int) ([]domain.Article, error)
	GetById(ctx context.Context, id int64) (domain.Article, error)
	GetPublishedById(ctx context.Context, id, uid int64) (domain.Article, error)
}

type ArticleServiceStruct struct {
	repo repository.ArticleRepository
}

func NewArticleServiceStruct(repo repository.ArticleRepository) ArticleService {
	return &ArticleServiceStruct{repo: repo}
}

func (s *ArticleServiceStruct) Save(ctx context.Context, art domain.Article) (int64, error) {
	art.Status = domain.ArticleStatusUnpublished
	if art.Id > 0 {
		err := s.repo.Update(ctx, art)
		return art.Id, err
	}
	return s.repo.Create(ctx, art)
}

func (s *ArticleServiceStruct) Withdraw(ctx context.Context, art domain.Article) error {
	art.Status = domain.ArticleStatusPrivate
	return s.repo.SyncStatus(ctx, art)
}

func (s *ArticleServiceStruct) Publish(ctx context.Context, art domain.Article) (int64, error) {
	art.Status = domain.ArticleStatusPublished
	// 制作库
	return s.repo.Sync(ctx, art)
}

func (s *ArticleServiceStruct) List(ctx context.Context, uid int64, offset int, limit int) ([]domain.Article, error) {
	return s.repo.List(ctx, uid, offset, limit)
}

func (s *ArticleServiceStruct) ListPub(ctx context.Context, start time.Time, offset, limit int) ([]domain.Article, error) {
	return s.repo.ListPub(ctx, start, offset, limit)
}

func (s *ArticleServiceStruct) GetById(ctx context.Context, id int64) (domain.Article, error) {
	return s.repo.GetById(ctx, id)
}

func (s *ArticleServiceStruct) GetPublishedById(ctx context.Context, id, uid int64) (domain.Article, error) {
	art, err := s.repo.GetPublishedById(ctx, id)
	//TODO implement me
	panic("implement me")
	//if err == nil {
	//	// 每次打开一篇文章，就发一条消息
	//	go func() {
	//		// 生产者也可以通过改批量来提高性能
	//		er := s.producer.ProduceReadEvent(
	//			ctx,
	//			events.ReadEvent{
	//				// 即便你的消费者要用 art 的里面的数据，
	//				// 让它去查询，你不要在 event 里面带
	//				Uid: uid,
	//				Aid: id,
	//			})
	//		if er != nil {
	//			svc.l.Error("发送读者阅读事件失败")
	//		}
	//	}()
	//
	//	//go func() {
	//	//	// 改批量的做法
	//	//	svc.ch <- readInfo{
	//	//		aid: id,
	//	//		uid: uid,
	//	//	}
	//	//}()
	//}
	return art, err
}
