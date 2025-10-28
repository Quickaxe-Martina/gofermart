package config

// Config variables
type Config struct {
	RunAddr              string `env:"SERVER_ADDRESS"`
	DatabaseDsn          string `env:"DATABASE_DSN"`
	AccuralSystemAddress string `env:"ACCRUAL_SYSTEM_ADDRESS"`
	MigrationsPath       string `env:"MIGRATIONS_PATH"`
	DevMode              bool   `env:"DEV_MODE"`
	SecretKey            string `env:"SECRET_KEY"`
	TokenExp             int    `env:"TOKEN_EXP"`
	DeleteBachSize       int    `env:"DELETE_BACH_SIZE"`
	DeleteTimeDuration   int    `env:"DELETE_TIME_DURATION"`
}

// NewConfig create Config
func NewConfig() *Config {
	var cfg = Config{
		RunAddr:              "",
		DatabaseDsn:          "",
		AccuralSystemAddress: "",
		MigrationsPath:       "./migrations",
		DevMode:              false,
		SecretKey:            "my_secret_key",
		TokenExp:             3,
		DeleteTimeDuration:   5,
		DeleteBachSize:       50,
	}
	LoadEnv(&cfg)
	ParseFlags(&cfg, true)
	return &cfg
}
