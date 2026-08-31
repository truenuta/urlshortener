package config

import (
	"flag"
	"time"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	Address              string        `env:"SERVER_ADDRESS"`
	BaseShortURLAddress  string        `env:"BASE_URL"`
	FileStoragePath      string        `env:"FILE_STORAGE_PATH"`
	DataBaseDSN          string        `env:"DATABASE_DSN"`
	DeleteWorkers        int           `env:"DELETE_WORKERS"`
	DeleteFlushThreshold int           `env:"DELETE_FLUSH_THRESHOLD"`
	DeleteFlushInterval  time.Duration `env:"DELETE_FLUSH_INTERVAL"`
	DeleteQueueCapacity  int           `env:"DELETE_QUEUE_CAPACITY"`
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
	if cfg.DataBaseDSN == "" {
		cfg.DataBaseDSN = cfgFlag.DataBaseDSN
	}
	if cfg.DeleteWorkers == 0 {
		cfg.DeleteWorkers = cfgFlag.DeleteWorkers
	}
	if cfg.DeleteFlushThreshold == 0 {
		cfg.DeleteFlushThreshold = cfgFlag.DeleteFlushThreshold
	}
	if cfg.DeleteFlushInterval == 0 {
		cfg.DeleteFlushInterval = cfgFlag.DeleteFlushInterval
	}
	if cfg.DeleteQueueCapacity == 0 {
		cfg.DeleteQueueCapacity = cfgFlag.DeleteQueueCapacity
	}

	return cfg, nil

}

func ParseFlags() *Config {
	var address string
	var baseShortURLAddress string
	var filePath string
	var DataBaseDSN string
	var deleteWorkers int
	var deleteFlushThreshold int
	var deleteFlushInterval time.Duration
	var deleteQueueCapacity int
	flag.StringVar(&address, "a", "localhost:8080", "адрес запуска HTTP-сервера")
	flag.StringVar(&baseShortURLAddress, "b", "http://localhost:8080", "базовый адрес результирующего сокращённого URL ")
	flag.StringVar(&filePath, "f", "storage.json", "путь к файлу хранения данных")
	flag.StringVar(&DataBaseDSN, "d", "", "строка подключения к базе данных")
	flag.IntVar(&deleteWorkers, "delete-workers", 4, "количество воркеров удаления URL")
	flag.IntVar(&deleteFlushThreshold, "delete-flush-threshold", 100, "порог сброса накопленных удалений")
	flag.DurationVar(&deleteFlushInterval, "delete-flush-interval", 5*time.Second, "интервал принудительного сброса")
	flag.IntVar(&deleteQueueCapacity, "delete-queue-capacity", 1024, "ёмкость очереди удаления")
	flag.Parse()
	return &Config{
		Address:              address,
		BaseShortURLAddress:  baseShortURLAddress,
		FileStoragePath:      filePath,
		DataBaseDSN:          DataBaseDSN,
		DeleteFlushInterval:  deleteFlushInterval,
		DeleteWorkers:        deleteWorkers,
		DeleteFlushThreshold: deleteFlushThreshold,
		DeleteQueueCapacity:  deleteQueueCapacity,
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
