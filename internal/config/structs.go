package config

// Config variables
type Config struct {
	RunAddr              string `env:"SERVER_ADDRESS"`
	DatabaseDsn          string `env:"DATABASE_DSN"`
	AccrualSystemAddress string `env:"ACCRUAL_SYSTEM_ADDRESS"`
	MigrationsPath       string `env:"MIGRATIONS_PATH"`
	SecretKey            string `env:"SECRET_KEY"`
	TokenExp             int    `env:"TOKEN_EXP"`
	NumWorkers           int    `env:"NUM_WORKERS"`
	PullSize             int    `env:"PULL_SIZE"`
	PoolTimeout          int    `env:"PULL_TIMEOUT"`
}

// NewConfig create Config
func NewConfig() *Config {
	var cfg = Config{
		RunAddr:              "",
		DatabaseDsn:          "",
		AccrualSystemAddress: "",
		MigrationsPath:       "./migrations",
		SecretKey:            "my_secret_key",
		TokenExp:             3,
		PullSize:             30,
		PoolTimeout:          5,
		NumWorkers:           3,
	}
	LoadEnv(&cfg)
	ParseFlags(&cfg, true)
	return &cfg
}
