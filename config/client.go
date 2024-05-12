package config

import (
	"github.com/ilyakaznacheev/cleanenv"
)

type ClientCfg struct {
	Host      string `env:"CLIENT_HOST" env-default:"localhost"`
	Port      string `env:"CLIENT_PORT" env-default:"8080"`
	Path      string `env:"CHAT_PATH" env-default:"/chat"`
	Scheme    string `env:"CLIENT_SCHEME" env-default:"ws"`
	LoggerLVL string `env:"CLIENT_LOGGER_LEVEL" env-default:"info"`
}

func NewClientCfg() (ClientCfg, error) {
	var res ClientCfg
	if err := cleanenv.ReadEnv(&res); err != nil {
		return ClientCfg{}, err
	}
	return res, nil
}
