package config

import (
	"github.com/ilyakaznacheev/cleanenv"
)

type ServerCfg struct {
	Server    ServerAddr
	Storage   StorageAddr
	Redis     RedisAddr
	Kafka     KafkaCfg
	LoggerLVL string `env:"SERVER_LOGGER_LEVEL" env-default:"info"`
}

type ServerAddr struct {
	Host string `env:"SRV_HOST" env-default:"localhost"`
	Port string `env:"SRV_PORT" env-default:"8080"`
}

type RedisAddr struct {
	Host       string `env:"REDIS_HOST" env-default:"localhost"`
	Port       string `env:"REDIS_PORT" env-default:"6379"`
	MaxRecords int64  `env:"REDIS_MAX_RECORDS" env-default:"1000"`
	HeadSize   int64  `env:"REDIS_HEAD_SIZE" env-default:"10"`
}

func NewServerCfg() (ServerCfg, error) {
	var res ServerCfg
	if err := cleanenv.ReadEnv(&res); err != nil {
		return ServerCfg{}, err
	}
	return res, nil
}
