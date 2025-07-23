package cache

import (
	"compus_blog/basic/internal/domain"
	"context"
	"encoding/json"
	"fmt"
	"github.com/redis/go-redis/v9"
	"time"
)

type ArticleCache interface {
	DelFirstPage(ctx context.Context, uid int64) error
	// GetFirstPage 只缓存第第一页的数据
	// 并且不缓存整个 Content
	GetFirstPage(ctx context.Context, uid int64) ([]domain.Article, error)
	SetFirstPage(ctx context.Context, uid int64, arts []domain.Article) error

	Set(ctx context.Context, art domain.Article) error
	Get(ctx context.Context, id int64) (domain.Article, error)

	SetPub(ctx context.Context, art domain.Article) error
	DelPub(ctx context.Context, id int64) error
	GetPub(ctx context.Context, id int64) (domain.Article, error)
}

type ArticleCacheStruct struct {
	client redis.Cmdable
}

func NewArticleCacheStruct(client redis.Cmdable) ArticleCache {
	return &ArticleCacheStruct{client: client}
}

func (a *ArticleCacheStruct) DelFirstPage(ctx context.Context, uid int64) error {
	return a.client.Del(ctx, a.firstPageKey(uid)).Err()
}

func (a *ArticleCacheStruct) GetFirstPage(ctx context.Context, uid int64) ([]domain.Article, error) {
	bs, err := a.client.Get(ctx, a.firstPageKey(uid)).Bytes()
	if err != nil {
		return nil, err
	}
	var arts []domain.Article
	err = json.Unmarshal(bs, &arts)
	return arts, err
}

func (a *ArticleCacheStruct) SetFirstPage(ctx context.Context, uid int64, arts []domain.Article) error {
	for i := range arts {
		//只缓存摘要部分
		arts[i].Content = arts[i].Abstract()
	}
	bs, err := json.Marshal(arts)
	if err != nil {
		return err
	}
	return a.client.Set(ctx, a.firstPageKey(uid), bs, time.Minute*10).Err()
}

func (a *ArticleCacheStruct) Set(ctx context.Context, art domain.Article) error {
	data, err := json.Marshal(art)
	if err != nil {
		return err
	}
	return a.client.Set(ctx, a.authorArtKey(art.Id), data, time.Hour).Err()
}

func (a *ArticleCacheStruct) Get(ctx context.Context, id int64) (domain.Article, error) {
	// 可以直接使用 Bytes 方法来获得 []byte
	data, err := a.client.Get(ctx, a.authorArtKey(id)).Bytes()
	if err != nil {
		return domain.Article{}, err
	}
	var art domain.Article
	err = json.Unmarshal(data, &art)
	return art, err
}

func (a *ArticleCacheStruct) SetPub(ctx context.Context, art domain.Article) error {
	data, err := json.Marshal(art)
	if err != nil {
		return err
	}
	return a.client.Set(ctx, a.readerArtKey(art.Id), data,
		// 设置过期时间
		time.Minute*30).Err()
}

func (a *ArticleCacheStruct) DelPub(ctx context.Context, id int64) error {
	return a.client.Del(ctx, a.readerArtKey(id)).Err()
}

func (a *ArticleCacheStruct) GetPub(ctx context.Context, id int64) (domain.Article, error) {
	// 可以直接使用 Bytes 方法来获得 []byte
	data, err := a.client.Get(ctx, a.readerArtKey(id)).Bytes()
	if err != nil {
		return domain.Article{}, err
	}
	var art domain.Article
	err = json.Unmarshal(data, &art)
	return art, err
}

func (a *ArticleCacheStruct) readerArtKey(id int64) string {
	return fmt.Sprintf("article_reader:%d", id)
}

func (a *ArticleCacheStruct) authorArtKey(id int64) string {
	return fmt.Sprintf("article_author:%d", id)
}

func (a *ArticleCacheStruct) firstPageKey(uid int64) string {
	return fmt.Sprintf("article:first_page:%d", uid)
}
