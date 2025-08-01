package domain

type Interactive struct {
	//	Biz+BizId定位文章
	Biz   string
	BizId int64

	// 阅读、点赞、收藏计数
	ReadCnt    int64 `json:"read_cnt"`
	LikeCnt    int64 `json:"like_cnt"`
	CollectCnt int64 `json:"collect_cnt"`

	// 点赞、收藏
	Liked     bool `json:"liked"`
	Collected bool `json:"collected"`
}
