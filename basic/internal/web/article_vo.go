package web

import "compus_blog/basic/internal/domain"

type ArticleVo struct {
	Id         int64  `json:"id,omitempty"`
	Title      string `json:"title,omitempty"`
	Abstract   string `json:"abstract,omitempty"`
	Content    string `json:"content,omitempty"`
	AuthorId   int64  `json:"authorId,omitempty"`
	AuthorName string `json:"authorName,omitempty"`
	Status     uint8  `json:"status,omitempty"`
	Ctime      string `json:"ctime,omitempty"`
	Utime      string `json:"utime,omitempty"`

	// 计数
	ReadCnt    int64 `json:"readCnt"`
	LikeCnt    int64 `json:"likeCnt"`
	CollectCnt int64 `json:"collectCnt"`

	// 我个人有没有收藏，有没有点赞
	Liked     bool `json:"liked"`
	Collected bool `json:"collected"`
}

type ArticleReq struct {
	Id      int64  `json:"id"`
	Title   string `json:"title"`
	Content string `json:"content"`
}

type CollectReq struct {
	Id      int64 `json:"id"`
	Collect bool  `json:"collect"`
}

type ListReq struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

type RewardReq struct {
	Id     int64 `json:"id"`
	Amount int64 `json:"amount"`
}

// VO view object，就是对标前端的

type LikeReq struct {
	Id int64 `json:"id"`
	// 点赞和取消点赞，我都准备复用这个
	Like bool `json:"like"`
}

func (req ArticleReq) toDomain(uid int64) domain.Article {
	return domain.Article{
		Id:      req.Id,
		Title:   req.Title,
		Content: req.Content,
		Author: domain.Author{
			Id: uid,
		},
	}
}
