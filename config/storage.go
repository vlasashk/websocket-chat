package config

import "github.com/ilyakaznacheev/cleanenv"

type StorageConfig struct {
	HTTP      StorageAddr
	Repo      RepoCfg
	Kafka     KafkaCfg
	LoggerLVL string `env:"STORAGE_LOGGER_LEVEL" env-default:"info"`
}

type RepoCfg struct {
	Schema        string `env:"DB_SCHEMA" env-default:"postgres"`
	Host          string `env:"DB_HOST" env-default:"localhost"`
	Port          string `env:"DB_PORT" env-default:"5432"`
	Name          string `env:"DB_NAME" env-default:"postgres"`
	User          string `env:"DB_USER" env-default:"postgres"`
	Password      string `env:"DB_PASSWORD" env-required:"true"`
	MigrationPath string `env:"DB_MIGRATION_PATH" env-default:"./migrations"`
}

type KafkaCfg struct {
	Topic     string `env:"KAFKA_TOPIC" env-default:"chat"`
	Partition int    `env:"KAFKA_PARTITION" env-default:"0"`
	Addr      string `env:"KAFKA_ADDR" env-default:"localhost:9092"`
	GroupID   string `env:"KAFKA_GROUP_ID" env-default:"chat"`
	BatchSize int    `env:"KAFKA_BATCH_SIZE" env-default:"10"`
	Async     bool   `env:"KAFKA_ASYNC" env-default:"true"`
}

type StorageAddr struct {
	Host string `env:"STORAGE_HOST" env-default:"localhost"`
	Port string `env:"STORAGE_PORT" env-default:"8000"`
}

func NewStorageCfg() (StorageConfig, error) {
	var res StorageConfig
	if err := cleanenv.ReadEnv(&res); err != nil {
		return StorageConfig{}, err
	}
	return res, nil
}
