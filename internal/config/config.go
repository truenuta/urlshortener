package config

import (
	"flag"
	"log"

	"github.com/caarlos0/env/v6"
)

type Config struct {
	Address             string `env:"SERVER_ADDRESS"`
	BaseShortURLAddress string `env:"BASE_URL"`
	FileStoragePath     string `env:"FILE_STORAGE_PATH"`
}

func NewConfig() *Config {
	cfgFlag := ParseFlags()
	cfg := ParseEnvList()
	if cfg.Address == "" {
		cfg.Address = cfgFlag.Address
	}
	if cfg.BaseShortURLAddress == "" {
		cfg.BaseShortURLAddress = cfgFlag.BaseShortURLAddress
	}
	if cfg.FileStoragePath == "" {
		cfg.FileStoragePath = cfgFlag.FileStoragePath
	}

	return cfg

}

func ParseFlags() *Config {
	var address string
	var baseShortURLAddress string
	var filePath string
	flag.StringVar(&address, "a", "localhost:8080", "адрес запуска HTTP-сервера")
	flag.StringVar(&baseShortURLAddress, "b", "http://localhost:8080", "базовый адрес результирующего сокращённого URL ")
	flag.StringVar(&filePath, "f", "storage.json", "путь к файлу хранения данных")
	flag.Parse()
	return &Config{
		Address:             address,
		BaseShortURLAddress: baseShortURLAddress,
		FileStoragePath:     filePath,
	}
}

func ParseEnvList() *Config {
	var cfg Config
	err := env.Parse(&cfg)
	if err != nil {
		log.Fatal(err)
	}
	return &cfg

}
