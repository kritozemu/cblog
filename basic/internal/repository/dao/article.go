package dao

import (
	"context"
	"errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

type ArticleDAOStruct struct {
	db *gorm.DB
}

func NewArticleDAOStruct(db *gorm.DB) ArticleDAO {
	return &ArticleDAOStruct{
		db: db,
	}
}

func (dao *ArticleDAOStruct) Insert(ctx context.Context, art Article) (int64, error) {
	now := time.Now().UnixMilli()
	art.Ctime = now
	art.Utime = now
	err := dao.db.WithContext(ctx).Create(&art).Error
	return art.Id, err
}

// UpdateById 只更新标题、内容和状态
func (dao *ArticleDAOStruct) UpdateById(ctx context.Context, art Article) error {
	now := time.Now().UnixMilli()
	res := dao.db.WithContext(ctx).Model(&Article{}).
		Where("id=? AND author_id = ?", art.Id, art.AuthorId).Updates(map[string]any{
		"Title":   art.Title,
		"content": art.Content,
		"status":  art.Status,
		"Utime":   now,
	})
	err := res.Error
	if err != nil {
		return err
	}
	if res.RowsAffected == 0 {
		return errors.New("更新数据失败")
	}
	return nil
}

func (dao *ArticleDAOStruct) SyncStatus(ctx context.Context, id int64, author int64, status uint8) error {
	return dao.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		res := tx.Model(&Article{}).
			Where("id=? AND author_id = ?", id, author).Update("status", status)
		err := res.Error
		if err != nil {
			return err
		}
		if res.RowsAffected != 1 {
			return ErrPossibleIncorrectAuthor
		}

		res = tx.Model(&PublishedArticle{}).
			Where("id=? AND author_id = ?", id, author).Update("status", status)
		err = res.Error
		if err != nil {
			return err
		}
		if res.RowsAffected != 1 {
			return ErrPossibleIncorrectAuthor
		}
		return nil

	})
}

func (dao *ArticleDAOStruct) Sync(ctx context.Context, art Article) (int64, error) {
	tx := dao.db.WithContext(ctx).Begin()
	now := time.Now().UnixMilli()
	defer tx.Rollback()
	// article
	txDAO := NewArticleDAOStruct(tx)
	var (
		id  = art.Id
		err error
	)

	if id == 0 {
		id, err = txDAO.Insert(ctx, art)
	} else {
		err = txDAO.UpdateById(ctx, art)
	}
	if err != nil {
		return 0, err
	}
	art.Id = id

	//publishedArticle
	pubArticle := PublishedArticle(art)
	pubArticle.Ctime = now
	pubArticle.Utime = now
	err = tx.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "id"}},
		DoUpdates: clause.Assignments(map[string]any{
			"title":   pubArticle.Title,
			"content": pubArticle.Content,
			"status":  pubArticle.Status,
			"Utime":   now,
		}),
	}).Create(pubArticle).Error
	if err != nil {
		return 0, err
	}
	tx.Commit()
	return id, tx.Error
}

func (dao *ArticleDAOStruct) GetByAuthor(ctx context.Context, uid int64, offset int, limit int) ([]Article, error) {
	var arts []Article
	err := dao.db.WithContext(ctx).Model(&Article{}).Where("author_id=?", uid).Offset(offset).
		Limit(limit).Order("utime desc").Find(&arts).Error
	return arts, err
}

func (dao *ArticleDAOStruct) ListPub(ctx context.Context, start time.Time, offset int, limit int) ([]Article, error) {
	var arts []Article
	err := dao.db.WithContext(ctx).Model(&Article{}).Where("utime < ?", start.UnixMilli()).
		Offset(offset).Limit(limit).Find(&arts).Error
	return arts, err
}

func (dao *ArticleDAOStruct) GetById(ctx context.Context, id int64) (Article, error) {
	var art Article
	err := dao.db.WithContext(ctx).Model(&Article{}).Where("id=?", id).First(&art).Error
	return art, err
}

func (dao *ArticleDAOStruct) GetPublishedById(ctx context.Context, id int64) (PublishedArticle, error) {
	var pub PublishedArticle
	err := dao.db.WithContext(ctx).Where("id=?", id).First(&pub).Error
	return pub, err
}
