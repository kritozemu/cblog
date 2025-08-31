//go:build wireinject

package main

import (
	grpc2 "compus_blog/basic/comment/grpc"
	"compus_blog/basic/comment/ioc"
	"compus_blog/basic/comment/repository"
	"compus_blog/basic/comment/repository/dao"
	"compus_blog/basic/comment/service"
	"github.com/google/wire"
)

var serviceProviderSet = wire.NewSet(
	dao.NewCommentDAO,
	repository.NewCommentRepo,
	service.NewCommentSvc,
	grpc2.NewGrpcServer,
)

var thirdProvider = wire.NewSet(
	ioc.InitLogger,
	ioc.InitDB,
)

func Init() *App {
	wire.Build(
		thirdProvider,
		serviceProviderSet,
		ioc.InitGRPCxServer,
		wire.Struct(new(App), "*"),
	)
	return new(App)
}
