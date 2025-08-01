package ginx

import (
	"compus_blog/basic/internal/web/jwt"
	"github.com/gin-gonic/gin"
)

func WrapReq[T any](fn func(ctx *gin.Context, req T, uc jwt.UserClaims) (Result, error)) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// 顺便把 userClaims 也取出来
	}
}

type Result struct {
	// 这个叫做业务错误码
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data any    `json:"data"`
}
