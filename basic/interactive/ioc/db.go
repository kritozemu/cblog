package ioc

import (
	"compus_blog/basic/interactive/repository/dao"
	"compus_blog/basic/pkg/logger"
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func InitSrc(l logger.LoggerV1) *gorm.DB {
	return InitDB(l, "src")
}

func InitDst(l logger.LoggerV1) DstDB {
	return InitDB(l, "dst")
}

type SrcDB *gorm.DB
type DstDB *gorm.DB

func InitDB(l logger.LoggerV1, key string) *gorm.DB {
	type Config struct {
		DSN string `yaml:"dsn"`
	}
	var cfg = Config{
		DSN: "root:123@tcp(localhost:13316)/cblog",
	}
	err := viper.UnmarshalKey("db"+key, &cfg)
	db, err := gorm.Open(mysql.Open(cfg.DSN), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	err = dao.InitTables(db)
	if err != nil {
		panic(err)
	}
	return db
}
