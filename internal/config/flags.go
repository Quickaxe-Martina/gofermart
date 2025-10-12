package config

import (
	"flag"
	"strings"
)

// ParseFlags parses command-line flags to configure the server's runtime parameters.
func ParseFlags(cfg *Config, onlyEmpty bool) {
	var runAddr string
	var serverAddr = "http://localhost:8080/"
	var dataFilePath string
	var databaseDsn string

	flag.StringVar(&runAddr, "a", ":8080", "address and port to run server")
	flag.StringVar(&dataFilePath, "f", "", "saved urls path")
	flag.StringVar(&databaseDsn, "d", "", "database")
	flag.Func("b", "server address before short URL", func(s string) error {
		if len(s) == 0 {
			return nil
		}
		if !strings.HasSuffix(s, "/") {
			s += "/"
		}
		serverAddr = s
		return nil
	})
	flag.Parse()
	if onlyEmpty && cfg.RunAddr == "" {
		cfg.RunAddr = runAddr
	}
	if onlyEmpty && cfg.ServerAddr == "" {
		cfg.ServerAddr = serverAddr
	}
	if onlyEmpty && cfg.DataFilePath == "" {
		cfg.DataFilePath = dataFilePath
	}
	if onlyEmpty && cfg.DatabaseDsn == "" {
		cfg.DatabaseDsn = databaseDsn
	}
}
