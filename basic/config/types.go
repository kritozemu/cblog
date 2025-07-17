package config

type config struct {
	DB    DBConfig
	Redis RedisConfig
}

type DBConfig struct {
	Dsn string
}

type RedisConfig struct {
	Addr string
}
