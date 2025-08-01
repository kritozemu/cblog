package ginx

import (
	logger2 "compus_blog/basic/pkg/logger"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/prometheus/client_golang/prometheus"
	"net/http"
	"strconv"
)

// L 使用包变量
var L logger2.LoggerV1

var vector *prometheus.CounterVec

func InitCounter(opt prometheus.CounterOpts) {
	// 可以考虑使用 code, method, 命中路由，HTTP 状态码
	vector = prometheus.NewCounterVec(opt,
		[]string{"code"})
	prometheus.MustRegister(vector)
}

func Wrap(fn func(ctx *gin.Context) (Result, error)) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		res, err := fn(ctx)
		if err != nil {
			// 开始处理 error，其实就是记录一下日志
			L.Error("处理业务逻辑出错",
				logger2.String("path", ctx.Request.URL.Path),
				// 命中的路由
				logger2.String("route", ctx.FullPath()),
				logger2.Error(err))
		}
		vector.WithLabelValues(strconv.Itoa(res.Code)).Inc()
		ctx.JSON(http.StatusOK, res)
	}
}

func WrapToken[C jwt.Claims](fn func(ctx *gin.Context, uc C) (Result, error)) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// 执行一些东西
		val, ok := ctx.Get("users")
		if !ok {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		c, ok := val.(C)
		if !ok {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// 我的业务逻辑有可能要操作 ctx
		// 读取 HTTP HEADER
		res, err := fn(ctx, c)
		if err != nil {
			// 开始处理 error，其实就是记录一下日志
			L.Error("处理业务逻辑出错",
				logger2.String("path", ctx.Request.URL.Path),
				// 命中的路由
				logger2.String("route", ctx.FullPath()),
				logger2.Error(err))
		}
		vector.WithLabelValues(strconv.Itoa(res.Code)).Inc()
		ctx.JSON(http.StatusOK, res)
	}
}

func WrapBodyAndToken[Req any, C jwt.Claims](fn func(ctx *gin.Context, req Req, uc C) (Result, error)) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req Req
		if err := ctx.Bind(&req); err != nil {
			return
		}

		val, ok := ctx.Get("users")
		if !ok {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		c, ok := val.(C)
		if !ok {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// 我的业务逻辑有可能要操作 ctx
		// 读取 HTTP HEADER
		res, err := fn(ctx, req, c)
		if err != nil {
			// 开始处理 error，其实就是记录一下日志
			L.Error("处理业务逻辑出错",
				logger2.String("path", ctx.Request.URL.Path),
				// 命中的路由
				logger2.String("route", ctx.FullPath()),
				logger2.Error(err))
		}
		ctx.JSON(http.StatusOK, res)
	}
}

func WrapBodyV1[T any](fn func(ctx *gin.Context, req T) (Result, error)) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req T
		if err := ctx.Bind(&req); err != nil {
			return
		}

		// 我的业务逻辑有可能要操作 ctx
		// 读取 HTTP HEADER
		res, err := fn(ctx, req)
		if err != nil {
			// 开始处理 error，其实就是记录一下日志
			L.Error("处理业务逻辑出错",
				logger2.String("path", ctx.Request.URL.Path),
				// 命中的路由
				logger2.String("route", ctx.FullPath()),
				logger2.Error(err))
		}
		ctx.JSON(http.StatusOK, res)
	}
}

func WrapBody[T any](l logger2.LoggerV1, fn func(ctx *gin.Context, req T) (Result, error)) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req T
		if err := ctx.Bind(&req); err != nil {
			return
		}

		// 我的业务逻辑有可能要操作 ctx
		// 要读取 HTTP HEADER
		res, err := fn(ctx, req)
		if err != nil {
			// 开始处理 error，其实就是记录一下日志
			l.Error("处理业务逻辑出错",
				logger2.String("path", ctx.Request.URL.Path),
				// 命中的路由
				logger2.String("route", ctx.FullPath()),
				logger2.Error(err))
		}
		ctx.JSON(http.StatusOK, res)
	}
}
