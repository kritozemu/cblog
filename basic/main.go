package main

import (
	"github.com/gin-gonic/gin"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"log"
	"net/http"
)

func main() {
	initViper()
	initLogger()
	server := InitWebServer()

	server.GET("/hello", func(c *gin.Context) {
		c.String(http.StatusOK, "hello world")
	})
	err := server.Run(":8080")
	if err != nil {
		return
	}
}

func initViper() {
	cfile := pflag.String("config", "config.yaml", "文件配置路径")
	pflag.Parse()
	viper.SetConfigType("yaml")
	viper.SetConfigFile(*cfile)
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}
	val := viper.Get("test")
	log.Println(val)
}

func initLogger() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	zap.ReplaceGlobals(logger)
}
