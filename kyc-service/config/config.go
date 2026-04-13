package config

import "github.com/ilyakaznacheev/cleanenv"

type Config struct {
	HTTP HTTPConfig `env-prefix:"HTTP_"`
	Log  LogConfig  `env-prefix:"LOG_"`
}

type HTTPConfig struct {
	Host string `env:"HOST" env-default:"0.0.0.0"`
	Port int    `env:"PORT" env-default:"8081"`
}

type LogConfig struct {
	Level string `env:"LEVEL" env-default:"info"`
}

func Load() (*Config, error) {
	var cfg Config
	if err := cleanenv.ReadEnv(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
