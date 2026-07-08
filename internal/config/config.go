package config

import (
	"flag"
)

type Config struct {
	Address             string
	BaseShortURLAddress string
}

func ParseFlags() *Config {
	var address string
	var baseShortURLAddress string
	flag.StringVar(&address, "a", "localhost:8080", "адрес запуска HTTP-сервера")
	flag.StringVar(&baseShortURLAddress, "b", "http://localhost:8080", "базовый адрес результирующего сокращённого URL ")
	flag.Parse()
	return &Config{
		Address:             address,
		BaseShortURLAddress: baseShortURLAddress,
	}
}
