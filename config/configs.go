package configs

import (
	"fmt"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env        string `yaml: "env" env:"ENV" env-default:"local" env-required:"true"`
	HTTPServer `yaml:"http_server"`
	Logger     LogConfig  `yaml:"logger"`
	Db         DbConfig   `yaml:"db"`
	CORS       CORSConfig `yaml:"cors"`
}

type LogConfig struct {
	Level  int    `yaml:"level" env-default:"1"`
	Format string `yaml:"format" env-default:"console"`
}

type DbConfig struct {
	Dsn string `yaml:"dsn" env-required:"true" env:"DB_DSN"`
}

type CORSConfig struct {
	AllowedOrigins []string `yaml:"allowed_origins"`
}

type HTTPServer struct {
	Address     string        `yaml:"address" env-default:"localhost:8082"`
	Timeout     time.Duration `yaml:"timeout" env-default:"4s"`
	IdleTimeout time.Duration `yaml:"idle_timeout" env-default:"60s"`
	User        string        `yaml:"user" env-required:"true"`
	Password    string        `yaml:"password" env-required:"true" env:"HTTP_SERVER_PASSWORD"`
}

func Init() *Config {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "config/local.yml"
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		panic(fmt.Sprintf("Config file %s does not exist", configPath))
	}

	var cfg Config

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		panic(fmt.Sprintf("Failed to read config file: %s", err))
	}

	return &cfg
}
