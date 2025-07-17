package config

var Config = config{
	DB: DBConfig{
		Dsn: "root:123@tcp(localhost:13316)/cblog",
	},
	Redis: RedisConfig{
		Addr: "localhost:6379",
	},
}
