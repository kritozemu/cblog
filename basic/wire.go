//go:build wireinject

package main

import (
	"compus_blog/basic/internal/ioc"
	"compus_blog/basic/internal/repository"
	"compus_blog/basic/internal/repository/cache"
	"compus_blog/basic/internal/repository/dao"
	"compus_blog/basic/internal/service"
	"compus_blog/basic/internal/web"
	ijwt "compus_blog/basic/internal/web/jwt"
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
)

func InitWebServer() *gin.Engine {
	wire.Build(
		//第三方服务
		ioc.InitDB, ioc.InitRedis,
		ioc.InitLogger,

		//dao
		dao.InitTables,
		dao.NewUserDAO,

		//cache
		cache.NewUserCache,

		//repository
		repository.NewUserRepository,

		//service
		service.NewUserService,

		//handler
		web.NewUserHandler,
		ijwt.NewRedisJWTHandler,
		ioc.InitGinMiddleWares,
		ioc.InitWebServer,
	)
	return gin.Default()
}
