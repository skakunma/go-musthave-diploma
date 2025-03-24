package config

import (
	"fmt"

	"github.com/skakunma/go-musthave-diploma-tpl/internal/storage"
	"go.uber.org/zap"
)

type Config struct {
	Store         storage.PostgresStorage
	Salt          string
	Sugar         *zap.SugaredLogger
	FlagForDB     string
	FlagAddress   string
	FlagAddressAS string
}

func NewConfig() (*Config, error) {
	cfg := &Config{}
	ParseFlags(cfg)
	store, err := storage.CreatePostgreStorage(cfg.FlagForDB)
	if err != nil {
		return nil, err
	}
	cfg.Store = *store

	logger, err := zap.NewDevelopment()
	if err != nil {
		return nil, fmt.Errorf("ошибка инициализации логгера: %w", err)
	}
	cfg.Sugar = logger.Sugar()

	salt := "random_salt_123"
	cfg.Salt = salt

	return cfg, nil
}
