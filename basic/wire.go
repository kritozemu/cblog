//go:build wireinject

package main

import (
	"compus_blog/basic/internal/events/article"
	"compus_blog/basic/internal/ioc"
	"compus_blog/basic/internal/repository"
	"compus_blog/basic/internal/repository/cache"
	"compus_blog/basic/internal/repository/dao"
	"compus_blog/basic/internal/service"
	"compus_blog/basic/internal/web"
	ijwt "compus_blog/basic/internal/web/jwt"
	"github.com/google/wire"
)

// interactive
var interactiveSvcProvider = wire.NewSet(dao.NewInteractiveDAO,
	repository.NewInteractiveRepository,
	cache.NewInteractiveCache,
	service.NewInteractiveService,
)

var thirdPartSet = wire.NewSet(
	//第三方服务
	ioc.InitDB, ioc.InitRedis,
	ioc.InitLogger,
	ioc.InitKafka,
)

func InitWebServer() *App {
	wire.Build(

		interactiveSvcProvider,
		thirdPartSet,

		// consumer
		article.NewKafkaProducer,
		article.NewInteractiveReadEventBatchConsumer,

		ioc.InitSyncProducer,
		ioc.NewConsumers,

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
		repository.NewArticleRepository,
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
		wire.Struct(new(App), "*"),
	)
	return new(App)
}
