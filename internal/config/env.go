package config

import (
	"log"

	"github.com/caarlos0/env/v11"
)

// LoadEnv parses ENV to configure the server's runtime parameters.
func LoadEnv(cfg *Config) {
	err := env.Parse(cfg)
	if err != nil {
		log.Fatal(err)
	}
}
