package ioc

import (
	grpc2 "compus_blog/basic/interactive/grpc"
	"compus_blog/basic/pkg/grpcx"
	"compus_blog/basic/pkg/logger"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

func InitGrpcxServer(l logger.LoggerV1, intrServer *grpc2.InteractiveServiceServer) *grpcx.Server {
	type Config struct {
		Port     int    `yaml:"port"`
		EtcdAddr string `yaml:"etcdAddr"`
		Name     string `yaml:"name"`
	}

	var cfg Config
	err := viper.UnmarshalKey("grpc.server", &cfg)
	if err != nil {
		panic(err)
	}

	server := grpc.NewServer()
	intrServer.Register(server)

	return &grpcx.Server{
		Server:   server,
		Port:     cfg.Port,
		EtcdAddr: cfg.EtcdAddr,
		L:        l,
		Name:     cfg.Name,
	}
}
