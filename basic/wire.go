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
		dao.NewUserDAO,
		dao.NewArticleDAOStruct,
		//cache
		cache.NewUserCache,
		cache.NewCodeCache,
		cache.NewArticleCacheStruct,
		//repository
		repository.NewUserRepository,
		repository.NewCodeRepository,
		repository.NewArticleRepositoryStruct,
		// 直接基于内存实现
		ioc.InitSMSService,

		//service
		service.NewUserService,
		service.NewCodeService,
		service.NewArticleServiceStruct,
		//handler
		web.NewUserHandler,
		web.NewArticleHandler,
		ijwt.NewRedisJWTHandler,
		ioc.InitGinMiddleWares,
		ioc.InitWebServer,
	)
	return gin.Default()
}
