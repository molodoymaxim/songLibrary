package config

import (
	"github.com/ilyakaznacheev/cleanenv"
	"log"
	"os"
	"time"
)

type Config struct {
	Env        string `yaml:"env" env-default:"local"`
	Database   `yaml:"db"`
	HttpServer `yaml:"HttpServer"`
}

type Database struct {
	Host     string `yaml:"host" env-default:"db"`
	User     string `yaml:"user" env-default:"postgres"`
	Password string `yaml:"password" env-default:"postgres"`
	Port     int    `yaml:"port" env-default:"5432"`
	Dbname   string `yaml:"dbname" env-default:"songLibrary"`
}

type HttpServer struct {
	Address     string        `yaml:"address" env-default:"8080"`
	Timeout     time.Duration `yaml:"timeout" env-default:"4"`
	IdleTimeout time.Duration `yaml:"idle_timeout" env-default:"60"`
}

func MustLoad() *Config {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		log.Fatal("CONFIG_PATH environment variable not set")
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatal("CONFIG_PATH does not exist")
	}

	var cfg Config

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatal("Error reading config file", err)
	}

	return &cfg
}
