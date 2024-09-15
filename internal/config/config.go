package config

import (
	goflag "flag"
	env "github.com/caarlos0/env/v6"
	flag "github.com/spf13/pflag"
	"log/slog"
	"os"
)

const (
	defaultCryptoKey = "please-set-in-env"
)

type Config struct {
	Addr              string `env:"RUN_ADDRESS"`
	RemoteAccrualAddr string `env:"ACCRUAL_SYSTEM_ADDRESS"`
	DatabaseURI       string `env:"DATABASE_URI"`
	CryptoKey         string `env:"KEY"`
}

var config *Config

// GetConfig returns copy of config
func GetConfig() Config {
	return *config
}

func init() {
	// Configure logger
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{AddSource: true})))

	config = &Config{}
	// Flags
	flag.StringVarP(&config.Addr, "runaddress", "a", "", "The address to listen on for HTTP requests")
	flag.StringVarP(&config.RemoteAccrualAddr, "accrualsystemaddress", "t", "", "The address to check order accruals on")
	flag.StringVarP(&config.DatabaseURI, "databasedsn", "d", "", "Postgres PGX DSN to use as DB storage.")
	flag.StringVarP(&config.CryptoKey, "key", "k", defaultCryptoKey, "Key for authorizations cryptography.")
	flag.CommandLine.AddGoFlagSet(goflag.CommandLine)
	flag.Parse()
	// pflag handles --help itself.

	// Env. variables. This should take over the command line. Bad practice as I know.
	if err := env.Parse(config); err != nil {
		slog.Error("error parsing env variables", "error", err)
		panic(err)
	}
}
