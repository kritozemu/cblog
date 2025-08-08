package ioc

import (
	"compus_blog/basic/internal/web"
	ijwt "compus_blog/basic/internal/web/jwt"
	"compus_blog/basic/internal/web/middleware"
	"compus_blog/basic/pkg/ginx"
	"compus_blog/basic/pkg/ginx/middlewares/ratelimit"
	"compus_blog/basic/pkg/limiter"
	"compus_blog/basic/pkg/logger"
	"context"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/redis/go-redis/v9"
	"time"
)

func InitWebServer(mdls []gin.HandlerFunc,
	userhdl *web.UserHandler,
	arthdl *web.ArticleHandler) *gin.Engine {
	server := gin.Default()
	server.Use(mdls...)
	userhdl.RegisterRoutes(server)
	arthdl.RegisterRoutes(server)
	return server
}

func InitGinMiddleWares(redisClient redis.Cmdable, hdl ijwt.Handler,
	l logger.LoggerV1) []gin.HandlerFunc {
	ginx.InitCounter(prometheus.CounterOpts{
		Namespace: "ice_juicy",
		Subsystem: "cblog",
		Name:      "biz_code",
		Help:      "统计业务错误码",
	})

	return []gin.HandlerFunc{
		cors.New(cors.Config{
			AllowOrigins: []string{"http://localhost:3000"},
			//AllowMethods:     []string{"PUT", "PATCH"},
			AllowHeaders:     []string{"Content-Type", "Authorization"},
			ExposeHeaders:    []string{"x-jwt-token", "x-refresh-token"},
			AllowCredentials: true,
			//AllowOriginFunc: func(origin string) bool {
			//	return origin == "http://localhost"
			//},
			MaxAge: 12 * time.Hour,
		}),
		middleware.NewLoginJwtMiddlewareBuilder(hdl).
			IgnorePath("/users/login").
			IgnorePath("/users/signup").
			IgnorePath("/users/login_sms/code/send").
			IgnorePath("/users/login_sms").
			IgnorePath("/hello").
			CheckLogin(),
		middleware.NewLogMiddleWareBuilder(func(ctx context.Context, al middleware.AccessLog) {
			l.Debug("", logger.Field{Key: "access", Value: al})
		}).AllowReqBody().AllowRespBody().Build(),

		ratelimit.NewBuilder(limiter.NewRedisSlidingWindowLimiter(redisClient, time.Second, 1000)).Build(),
	}
}
