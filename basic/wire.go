//go:build wireinject

package main

import (
	"compus_blog/basic/internal/events/article"
	"compus_blog/basic/internal/repository"
	"compus_blog/basic/internal/repository/cache"
	"compus_blog/basic/internal/repository/dao"
	"compus_blog/basic/internal/service"
	"compus_blog/basic/internal/web"
	ijwt "compus_blog/basic/internal/web/jwt"
	"compus_blog/basic/ioc"
	"github.com/google/wire"
)

var thirdPartSet = wire.NewSet(
	//第三方服务
	ioc.InitDB, ioc.InitRedis,
	ioc.InitEtcd,
	ioc.InitLogger,
	ioc.InitKafka,
	ioc.InitSyncProducer,
)

var rankingServiceSet = wire.NewSet(
	repository.NewRankingRepository,
	cache.NewRankingRedisCache,
	cache.NewRankingLocalCache,
	service.NewBatchRankingService,
)

func InitWebServer() *App {
	wire.Build(

		thirdPartSet,

		//jobs
		ioc.InitRLockClient,
		rankingServiceSet,
		ioc.InitRankingJob,
		ioc.InitJobs,

		ioc.NewConsumers,

		ioc.InitIntrClientV1,
		// events
		article.NewKafkaProducer,

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
