package ioc

import (
	logger2 "compus_blog/basic/pkg/logger"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func InitLogger() logger2.LoggerV1 {
	cfg := zap.NewDevelopmentConfig()
	err := viper.Unmarshal(&cfg)
	if err != nil {
		panic(err)
	}
	log, err := cfg.Build()
	if err != nil {
		panic(err)
	}
	return logger2.NewZapLogger(log)
}
