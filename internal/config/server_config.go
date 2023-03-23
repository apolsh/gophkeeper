package config

import (
	"flag"

	"github.com/caarlos0/env"
)

const (
	PostgresStorageType = "postgres"
)

// ServerConfig configuration files for backend
type ServerConfig struct {
	DatabaseDSN    string `env:"DATABASE_DSN" envDefault:"postgresql://gophkeeper:gophkeeper@localhost:5432/gophkeeper"`
	ServerAddr     string `env:"SERVER_ADDRESS" envDefault:":3333"`
	Storage        string `env:"STORAGE" envDefault:"postgres"`
	LogLevel       string `env:"LOG_LEVEL" envDefault:"info"`
	TokenSecretKey string `env:"TOKEN_SECRET_KEY" envDefault:"secret"`
}

func (c *ServerConfig) populateEmptyFields(another ServerConfig) {
	if c.DatabaseDSN == "" && another.DatabaseDSN != "" {
		c.DatabaseDSN = another.DatabaseDSN
	}
	if c.ServerAddr == "" && another.ServerAddr != "" {
		c.ServerAddr = another.ServerAddr
	}
	if c.Storage == "" && another.Storage != "" {
		c.Storage = another.Storage
	}
	if c.TokenSecretKey == "" && another.TokenSecretKey != "" {
		c.TokenSecretKey = another.TokenSecretKey
	}
}

// LoadServerConfig reads environment variables and flags, prior to flags
func LoadServerConfig() ServerConfig {

	var mainConfig ServerConfig

	flag.StringVar(&mainConfig.DatabaseDSN, "d", "", "database DSN")
	flag.StringVar(&mainConfig.ServerAddr, "a", "", "GRPC server address")
	flag.StringVar(&mainConfig.Storage, "st", "", "storage type (postgres)")
	flag.StringVar(&mainConfig.TokenSecretKey, "s", "", "secret key for token generator")

	flag.Parse()

	var envsConfig ServerConfig
	if err := env.Parse(&envsConfig); err != nil {
		panic(err)
	}

	mainConfig.populateEmptyFields(envsConfig)

	return mainConfig
}
