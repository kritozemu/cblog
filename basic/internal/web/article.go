package web

import (
	intrv1 "compus_blog/basic/api/proto/gen/intr/v1"
	"compus_blog/basic/internal/domain"
	"compus_blog/basic/internal/service"
	"compus_blog/basic/internal/web/jwt"
	"compus_blog/basic/pkg/ginx"
	"compus_blog/basic/pkg/logger"
	"fmt"
	"github.com/ecodeclub/ekit/slice"
	"github.com/gin-gonic/gin"
	"golang.org/x/sync/errgroup"
	"net/http"
	"strconv"
	"time"
)

var _ handler = (*UserHandler)(nil)

type ArticleHandler struct {
	svc     service.ArticleService
	intrSvc intrv1.InteractiveServiceClient
	l       logger.LoggerV1
	biz     string
}

func NewArticleHandler(svc service.ArticleService, l logger.LoggerV1,
	intrSvc intrv1.InteractiveServiceClient) *ArticleHandler {
	return &ArticleHandler{
		svc:     svc,
		l:       l,
		intrSvc: intrSvc,
		biz:     "article",
	}
}

func (h *ArticleHandler) RegisterRoutes(server *gin.Engine) {
	ag := server.Group("/articles")
	ag.POST("/edit", h.Edit)
	ag.POST("/withdraw", h.Withdraw)
	ag.POST("/publish", h.Publish)

	// 创作者接口
	ag.GET("/detail/:id", ginx.WrapToken[jwt.UserClaims](h.Detail))
	// 按照道理来说，这边就是 GET 方法
	// /list?offset=?&limit=?
	ag.POST("/list", ginx.WrapBodyAndToken[ListReq, jwt.UserClaims](h.List))

	pub := ag.Group("/pub")
	pub.GET("/:id", h.PubDetail)
	// 传入一个参数，true 就是点赞, false 就是不点赞
	pub.POST("/like", ginx.WrapBodyAndToken[LikeReq, jwt.UserClaims](h.Like))
	pub.POST("/collect", ginx.WrapBodyAndToken[CollectReq, jwt.UserClaims](h.Collect))
}

func (h *ArticleHandler) Edit(ctx *gin.Context) {
	var req ArticleReq
	if err := ctx.Bind(&req); err != nil {
		return
	}
	c := ctx.MustGet("users")
	claims, ok := c.(jwt.UserClaims)
	if !ok {
		// 你可以考虑监控住这里
		//ctx.AbortWithStatus(http.StatusUnauthorized)
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		h.l.Error("未发现用户的 session 信息")
		return
	}

	// 检测输入，跳过这一步
	// 调用 svc 的代码
	id, err := h.svc.Save(ctx, req.toDomain(claims.Uid))
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		// 打日志？
		h.l.Error("保存帖子失败", logger.Error(err))
		return
	}
	ctx.JSON(http.StatusOK, Result{
		Data: id,
	})
}

func (h *ArticleHandler) Withdraw(ctx *gin.Context) {
	type Req struct {
		Id int64 `json:"id"`
	}
	var req Req
	if err := ctx.Bind(&req); err != nil {
		return
	}

	c := ctx.MustGet("users")
	claims, ok := c.(jwt.UserClaims)
	if !ok {
		// 你可以考虑监控住这里
		//ctx.AbortWithStatus(http.StatusUnauthorized)
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		h.l.Error("未发现用户的 session 信息")
		return
	}

	err := h.svc.Withdraw(ctx, domain.Article{
		Id: req.Id,
		Author: domain.Author{
			Id: claims.Uid,
		},
	})
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		// 打日志？
		h.l.Error("保存帖子失败", logger.Error(err))
		return
	}
	ctx.JSON(http.StatusOK, Result{
		Msg: "OK",
	})

}

func (h *ArticleHandler) Publish(ctx *gin.Context) {
	var req ArticleReq
	if err := ctx.Bind(&req); err != nil {
		return
	}
	c := ctx.MustGet("users")
	claims, ok := c.(jwt.UserClaims)
	if !ok {
		// 你可以考虑监控住这里
		//ctx.AbortWithStatus(http.StatusUnauthorized)
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		h.l.Error("未发现用户的 session 信息")
		return
	}

	id, err := h.svc.Publish(ctx, req.toDomain(claims.Uid))
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		// 打日志？
		h.l.Error("发表帖子失败", logger.Error(err))
		return
	}
	ctx.JSON(http.StatusOK, Result{
		Data: id,
	})
}

func (h *ArticleHandler) Detail(ctx *gin.Context, usr jwt.UserClaims) (ginx.Result, error) {
	idstr := ctx.Param("id")
	id, err := strconv.ParseInt(idstr, 10, 64)
	if err != nil {
		//ctx.JSON(http.StatusOK, )
		//a.l.Error("前端输入的 ID 不对", logger.Error(err))
		return ginx.Result{
			Code: 4,
			Msg:  "参数错误",
		}, err
	}
	art, err := h.svc.GetById(ctx, id)
	if err != nil {
		//ctx.JSON(http.StatusOK, )
		//a.l.Error("获得文章信息失败", logger.Error(err))
		return ginx.Result{
			Code: 5,
			Msg:  "系统错误",
		}, err
	}
	if art.Author.Id != usr.Uid {
		//ctx.JSON(http.StatusOK)
		// 如果有风控系统，这个时候就要上报这种非法访问的用户了。
		//a.l.Error("非法访问文章，创作者 ID 不匹配",
		//	logger.Int64("uid", usr.Id))
		return ginx.Result{
			Code: 4,
			// 也不需要告诉前端究竟发生了什么
			Msg: "输入有误",
		}, fmt.Errorf("非法访问文章，创作者 ID 不匹配 %d", usr.Uid)
	}
	return ginx.Result{
		Data: ArticleVo{
			Id:      art.Id,
			Title:   art.Title,
			Content: art.Content,
			// 不需要这个摘要信息
			//Abstract: art.Abstract(),
			Status: art.Status.ToUint8(),
			// 这个是创作者看自己的文章列表，也不需要这个字段
			AuthorId: art.Author.Id,
			Ctime:    art.Ctime.Format(time.DateTime),
			Utime:    art.Utime.Format(time.DateTime),
		},
	}, nil
}

func (h *ArticleHandler) List(ctx *gin.Context, req ListReq, uc jwt.UserClaims) (ginx.Result, error) {
	arts, err := h.svc.List(ctx, uc.Uid, req.Offset, req.Limit)
	if err != nil {
		return ginx.Result{
			Code: 5,
			Msg:  "系统错误",
		}, nil
	}
	return ginx.Result{
		Data: slice.Map[domain.Article, ArticleVo](arts, func(idx int, src domain.Article) ArticleVo {
			return ArticleVo{
				Id:       src.Id,
				Title:    src.Title,
				Abstract: src.Abstract(),

				Status: src.Status.ToUint8(),
				// 这个列表请求，不需要返回内容
				Ctime: src.Ctime.Format(time.DateTime),
				Utime: src.Utime.Format(time.DateTime),
			}
		}),
	}, nil
}

func (h *ArticleHandler) PubDetail(ctx *gin.Context) {
	idstr := ctx.Param("id")
	id, err := strconv.ParseInt(idstr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 4,
			Msg:  "参数错误",
		})
		h.l.Error("前端输入的 ID 不对", logger.String("idstr", idstr), logger.Error(err))
		return
	}

	uc := ctx.MustGet("users").(jwt.UserClaims)

	var (
		eg   errgroup.Group
		art  domain.Article
		intr *intrv1.GetResponse
	)

	eg.Go(func() error {
		art, err = h.svc.GetPublishedById(ctx, id, uc.Uid)
		return err
	})

	eg.Go(func() error {
		//intr, err = h.intrSvc.Get(ctx, h.biz, id, uc.Uid)
		intr, err = h.intrSvc.Get(ctx, &intrv1.GetRequest{
			Biz: h.biz, BizId: id, Uid: uc.Uid,
		})
		return err

	})

	//在这儿等，要保证前面两个
	err = eg.Wait()
	if err != nil {
		// 代表查询出错了
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}

	//go func() {
	//	er := h.intrSvc.IncrReadCnt(ctx, h.biz, id)
	//	if er != nil {
	//		h.l.Error("增加阅读计数失败",
	//			logger.Int64("aid", id),
	//			logger.Error(err))
	//	}
	//}()

	ctx.JSON(http.StatusOK, Result{
		Data: ArticleVo{
			Id:      art.Id,
			Title:   art.Title,
			Content: art.Content,
			Status:  art.Status.ToUint8(),

			//要把作者信息带出来
			AuthorId:   art.Author.Id,
			AuthorName: art.Author.Name,
			Ctime:      art.Ctime.Format(time.DateTime),
			Utime:      art.Utime.Format(time.DateTime),

			ReadCnt:    intr.Intr.ReadCnt,
			CollectCnt: intr.Intr.CollectCnt,
			LikeCnt:    intr.Intr.LikeCnt,
			Liked:      intr.Intr.Liked,
			Collected:  intr.Intr.Collected,
		},
	})

}

func (h *ArticleHandler) Like(ctx *gin.Context, req LikeReq, uc jwt.UserClaims) (ginx.Result, error) {
	var err error
	if req.Like {
		// 点赞
		_, err = h.intrSvc.Like(ctx, &intrv1.LikeRequest{
			Biz: h.biz, BizId: req.Id, Uid: uc.Uid,
		})
	} else {
		// 取消点赞
		_, err = h.intrSvc.CancelLike(ctx, &intrv1.CancelLikeRequest{
			Biz: h.biz, BizId: req.Id, Uid: uc.Uid,
		})
	}

	if err != nil {
		return ginx.Result{
			Code: 5,
			Msg:  "系统错误",
		}, err
	}
	return ginx.Result{Msg: "OK"}, nil
}

func (h *ArticleHandler) Collect(ctx *gin.Context, req CollectReq, uc jwt.UserClaims) (ginx.Result, error) {
	_, err := h.intrSvc.Collect(ctx, &intrv1.CollectRequest{
		Biz: h.biz, BizId: req.Id, Uid: uc.Uid, Cid: req.Cid,
	})

	if err != nil {
		return ginx.Result{
			Code: 5,
			Msg:  "系统错误",
		}, err
	}
	return ginx.Result{Msg: "OK"}, nil
}
