package main

import (
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"log"
	"net/http"
)

func main() {
	initViper()
	initLogger()
	initPrometheus()
	app := InitWebServer()
	for _, c := range app.consumers {
		err := c.Start()
		if err != nil {
			panic(err)
		}
	}

	app.server.Run(":8080")
	//server.GET("/hello", func(c *gin.Context) {
	//	c.String(http.StatusOK, "hello world")
	//})
	//err := server.Run(":8080")
	//if err != nil {
	//	return
	//}
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

func initPrometheus() {
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		http.ListenAndServe(":8081", nil)
	}()
}
