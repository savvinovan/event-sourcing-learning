package config

import "github.com/ilyakaznacheev/cleanenv"

type Config struct {
	HTTP  HTTPConfig  `env-prefix:"HTTP_"`
	Log   LogConfig   `env-prefix:"LOG_"`
	DB    DBConfig    `env-prefix:"DB_"`
	Kafka KafkaConfig `env-prefix:"KAFKA_"`
}

type HTTPConfig struct {
	Host string `env:"HOST" env-default:"0.0.0.0"`
	Port int    `env:"PORT" env-default:"8080"`
}

type LogConfig struct {
	Level string `env:"LEVEL" env-default:"info"`
}

type DBConfig struct {
	DSN string `env:"DSN" env-default:"postgres://wallet:wallet@192.168.1.61:5434/wallet?sslmode=disable"`
}

type KafkaConfig struct {
	Brokers string `env:"BROKERS" env-default:"192.168.1.61:9092"`
}

func Load() (*Config, error) {
	var cfg Config
	if err := cleanenv.ReadEnv(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
