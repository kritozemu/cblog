//go:build wireinject

package main

import (
	"compus_blog/basic/interactive/events"
	"compus_blog/basic/interactive/grpc"
	"compus_blog/basic/interactive/ioc"
	"compus_blog/basic/interactive/repository"
	"compus_blog/basic/interactive/repository/cache"
	"compus_blog/basic/interactive/repository/dao"
	"compus_blog/basic/interactive/service"
	"github.com/google/wire"
)

var thirdPartySet = wire.NewSet(
	//ioc.InitDst,
	ioc.InitSrc,
	ioc.InitLogger,
	ioc.InitKafka,

	ioc.InitSyncProducer,
	ioc.InitRedis,
)

// interactive
var interactiveSvcProvider = wire.NewSet(dao.NewInteractiveDAO,
	repository.NewInteractiveRepository,
	cache.NewInteractiveCache,
	service.NewInteractiveService,
)

func InitAPP() *App {
	wire.Build(
		interactiveSvcProvider,
		thirdPartySet,
		events.NewInteractiveReadEventBatchConsumer,
		ioc.NewConsumers,
		grpc.NewInteractiveServiceServer,
		ioc.InitGrpcxServer,
		wire.Struct(new(App), "*"),
	)
	return new(App)
}
