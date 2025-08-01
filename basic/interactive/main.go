package main

import (
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"log"
)

func main() {
	initViperV1()
	app := InitAPP()
	for _, c := range app.consumers {
		err := c.Start()
		if err != nil {
			panic(err)
		}
	}
	//go func() {
	//	err := app.webAdmin.Start()
	//	log.Println(err)
	//}()
	err := app.server.Serve()
	if err != nil {
		panic(err)
	}
}

func initViperV1() {
	cfile := pflag.String("config",
		"config/config.yaml", "配置文件路径")
	// 这一步之后，cfile 里面才有值
	pflag.Parse()
	//viper.Set("db.dsn", "localhost:3306")
	// 所有的默认值放好s
	viper.SetConfigType("yaml")
	viper.SetConfigFile(*cfile)
	// 读取配置
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}
	val := viper.Get("test.key")
	log.Println(val)
}
