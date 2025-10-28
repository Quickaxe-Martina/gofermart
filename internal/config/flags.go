package config

import (
	"flag"
)

// ParseFlags parses command-line flags to configure the server's runtime parameters.
func ParseFlags(cfg *Config, onlyEmpty bool) {
	var runAddr string
	var databaseDsn string
	var accuralSystemAddress string

	flag.StringVar(&runAddr, "a", ":8080", "address and port to run server")
	flag.StringVar(&databaseDsn, "d", "", "database")
	flag.StringVar(&accuralSystemAddress, "r", "", "accural system address")

	flag.Parse()
	if onlyEmpty && cfg.RunAddr == "" {
		cfg.RunAddr = runAddr
	}
	if onlyEmpty && cfg.DatabaseDsn == "" {
		cfg.DatabaseDsn = databaseDsn
	}
	if onlyEmpty && cfg.AccuralSystemAddress == "" {
		cfg.AccuralSystemAddress = accuralSystemAddress
	}
}
