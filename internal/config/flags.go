package config

import (
	"flag"
	"os"
)

func ParseFlags(cfg *Config) {
	// Получаем значения из переменных окружения (если они есть)
	envAddress := os.Getenv("RUN_ADDRESS")
	envDB := os.Getenv("DATABASE_URI")
	envAccrual := os.Getenv("ACCRUAL_SYSTEM_ADDRESS")

	// Устанавливаем флаги с учётом значений из окружения
	flag.StringVar(&cfg.FlagAddress, "a", ":8000", "address and port to run server")
	flag.StringVar(&cfg.FlagForDB, "d", "host=localhost user=postgres password=example dbname=diplomka sslmode=disable", "database conn link")
	flag.StringVar(&cfg.FlagAddressAS, "r", "http://localhost:8000", "address accrual system")

	flag.Parse() // Парсим флаги (они могут переопределить значения из окружения)

	if envAddress != "" {
		cfg.FlagAddress = envAddress
	}
	if envDB != "" {
		cfg.FlagForDB = envDB
	}

	if envAccrual != "" {
		cfg.FlagAddressAS = envAccrual
	}
}
