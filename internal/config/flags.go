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
	flag.StringVar(&cfg.FlagAddress, "a", envOrDefault(envAddress, ":8000"), "address and port to run server")
	flag.StringVar(&cfg.FlagForDB, "d", envOrDefault(envDB, ""), "database conn link")
	flag.StringVar(&cfg.FlagAddressAS, "r", envOrDefault(envAccrual, "http://localhost:8080"), "address accrual system")

	flag.Parse() // Парсим флаги (они могут переопределить значения из окружения)
}

func envOrDefault(envValue, defaultValue string) string {
	if envValue != "" {
		return envValue
	}
	return defaultValue
}
