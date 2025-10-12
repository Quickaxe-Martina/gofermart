package config

import (
	"log"

	"strings"

	"github.com/caarlos0/env/v11"
)

// LoadEnv parses ENV to configure the server's runtime parameters.
func LoadEnv(cfg *Config) {
	err := env.Parse(cfg)
	if err != nil {
		log.Fatal(err)
	}
	if len(cfg.ServerAddr) != 0 && !strings.HasSuffix(cfg.ServerAddr, "/") {
		cfg.ServerAddr += "/"
	}
}
