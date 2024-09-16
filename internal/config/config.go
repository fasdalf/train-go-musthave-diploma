package config

import (
	goflag "flag"
	env "github.com/caarlos0/env/v6"
	"github.com/fasdalf/train-go-musthave-diploma/internal/accrual"
	flag "github.com/spf13/pflag"
	"log/slog"
	"os"
	"time"
)

const (
	defaultCryptoKey = "please-set-in-env"
)

type Config struct {
	Addr        string        `env:"RUN_ADDRESS"`
	DatabaseURI string        `env:"DATABASE_URI"`
	CryptoKey   string        `env:"KEY"`
	TokenExp    time.Duration `env:"TOKEN_EXP"`
	Accrual     accrual.Config
}

// NewConfig creates copy of config
func NewConfig() (*Config, error) {
	//config := Config{Accrual: &accrual.Config{}}
	config := Config{}
	// Flags
	flag.StringVarP(&config.Addr, "runaddress", "a", "", "The address to listen on for HTTP requests")
	flag.StringVarP(&config.DatabaseURI, "databasedsn", "d", "", "Postgres PGX DSN to use as DB storage.")
	flag.StringVarP(&config.CryptoKey, "key", "k", defaultCryptoKey, "Key for authorizations cryptography.")
	flag.DurationVarP(&config.TokenExp, "tokenexp", "e", 3*time.Hour, "JWT token lifetime")
	flag.StringVarP(&config.Accrual.URL, "accrualsystemaddress", "r", "", "The address to check order accruals on")
	flag.IntVarP(&config.Accrual.WorkersCount, "accrualworkerscount", "w", 3, "number of background workers")
	flag.DurationVarP(&config.Accrual.FetchTimeout, "accrualfetchtimeout", "f", 300*time.Millisecond, "fetch HTTP timeout")
	flag.DurationVarP(&config.Accrual.FetchInterval, "accrualfetchinterval", "i", 400*time.Millisecond, "fetch timeout")
	flag.IntVarP(&config.Accrual.FetchFactor, "accrualfetchfactor", "m", 10, "factor to timeout to consider jobs stale")
	flag.CommandLine.AddGoFlagSet(goflag.CommandLine)
	flag.Parse()
	// pflag handles --help itself.

	// Env. variables. This should take over the command line. Bad practice as I know.
	if err := env.Parse(&config); err != nil {
		slog.Error("error parsing env variables", "error", err)
		return nil, err
	}
	return &config, nil
}

func init() {
	// Configure logger
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{AddSource: true})))
}
