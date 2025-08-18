//go:build wireinject

package startup

import (
	"compus_blog/basic/interactive/grpc"
	"compus_blog/basic/interactive/repository"
	"compus_blog/basic/interactive/repository/cache"
	"compus_blog/basic/interactive/repository/dao"
	"compus_blog/basic/interactive/service"
	"github.com/google/wire"
)

var thirdProvider = wire.NewSet(InitRedis,
	InitTestDB, InitLog)
var interactiveSvcProvider = wire.NewSet(
	service.NewInteractiveService,
	repository.NewInteractiveRepository,
	dao.NewInteractiveDAO,
	cache.NewInteractiveCache,
)

func InitInteractiveService() service.InteractiveService {
	wire.Build(thirdProvider, interactiveSvcProvider)
	return service.NewInteractiveService(nil, nil)
}

func InitInteractiveGRPCServer() *grpc.InteractiveServiceServer {
	wire.Build(thirdProvider, interactiveSvcProvider, grpc.NewInteractiveServiceServer)
	return new(grpc.InteractiveServiceServer)
}
