package config

import (
	"flag"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	Address             string `env:"SERVER_ADDRESS"`
	BaseShortURLAddress string `env:"BASE_URL"`
	FileStoragePath     string `env:"FILE_STORAGE_PATH"`
}

func NewConfig() (*Config, error) {
	cfgFlag := ParseFlags()
	cfg, err := ParseEnvList()
	if err != nil {
		return nil, err
	}
	if cfg.Address == "" {
		cfg.Address = cfgFlag.Address
	}
	if cfg.BaseShortURLAddress == "" {
		cfg.BaseShortURLAddress = cfgFlag.BaseShortURLAddress
	}
	if cfg.FileStoragePath == "" {
		cfg.FileStoragePath = cfgFlag.FileStoragePath
	}

	return cfg, nil

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

func ParseEnvList() (*Config, error) {
	var cfg Config
	err := env.Parse(&cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil

}
