package config

import "flag"

func ParseFlags(cfg *Config) {
	flag.StringVar(&cfg.FlagAddress, "a", ":8000", "address and port to run server")
	flag.StringVar(&cfg.FlagForDB, "d", "", "database conn link")
	flag.StringVar(&cfg.FlagAddressAS, "r", "http://localhost:8080", "address accrual system")
	flag.Parse()
}
