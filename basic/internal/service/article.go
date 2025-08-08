package service

import (
	"compus_blog/basic/internal/domain"
	"compus_blog/basic/internal/events/article"
	"compus_blog/basic/internal/repository"
	"compus_blog/basic/pkg/logger"
	"context"
	"time"
)

//go:generate mockgen -source=article.go -package=svcmocks -destination=mocks/article.mock.go ArticleService
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
	repo     repository.ArticleRepository
	producer article.Producer
	l        logger.LoggerV1
	ch       chan article.ReadEvent
}

func NewArticleServiceStruct(repo repository.ArticleRepository,
	producer article.Producer, l logger.LoggerV1) ArticleService {
	return &ArticleServiceStruct{
		repo:     repo,
		producer: producer,
		l:        l,
	}
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
	if err == nil {
		// 每次打开一篇文章，就发一条消息
		go func() {
			// 生产者也可以通过改批量来提高性能
			er := s.producer.ProduceReadEvent(
				ctx, article.ReadEvent{
					// 即便你的消费者要用 art 的里面的数据，
					// 让它去查询，你不要在 event 里面带
					Uid: uid,
					Aid: id,
				})
			if er != nil {
				s.l.Error("发送读者阅读事件失败")
			}
		}()

		//go func() {
		//	// 改批量的做法
		//	svc.ch <- readInfo{
		//		aid: id,
		//		uid: uid,
		//	}
		//}()
	}
	return art, err
}

func (s *ArticleServiceStruct) GetPublishedByIdV1(ctx context.Context, id, uid int64) (domain.Article, error) {
	// 另一个选项，在这里组装 Author，调用 UserService
	art, err := s.repo.GetPublishedById(ctx, id)
	if err == nil {
		go func() {
			// 改批量的做法
			s.ch <- article.ReadEvent{
				Uid: uid,
				Aid: id,
			}
		}()
	}
	return art, err
}

func (s *ArticleServiceStruct) batchSendReadInfo(ctx context.Context) {
	// 10 个一批
	// 单个转批量都要考虑的兜底问题
	for {
		ctx, cancel := context.WithTimeout(ctx, time.Second)
		const batchSize = 10
		events := make([]article.ReadEvent, 0, batchSize)
		send := false
		for !send {
			select {
			// 这边是超时了
			case <-ctx.Done():
				// 也要执行发送
				//goto send
				send = true
			case info, ok := <-s.ch:
				if !ok {
					cancel()
					send = true
					continue
				}
				events = append(events, info)
				// 凑够了
				if len(events) == batchSize {
					//goto send
					send = true
				}
			}
		}
		//send:
		// 装满了，凑够了一批
		s.producer.BatchProduceReadEventV1(context.Background(), events)
		cancel()
	}
}
