package config

import (
	"flag"
	"os"
	"path/filepath"

	"github.com/caarlos0/env"
)

// ClientConfig configuration files for backend.
type ClientConfig struct {
	BaseDir       string `env:"GOPHKEEPER_BASE_DIR"`
	SyncServerURL string `env:"GOPHKEEPER_SYNC_SERVER_URL" envDefault:":3333"`
	LogLevel      string `env:"GOPHKEEPER_LOG_LEVEL" envDefault:"info"`
	SyncPeriod    int64  `env:"GOPHKEEPER_SYNC_PERIOD" envDefault:"100"`
	HTTPSEnabled  bool   `env:"ENABLE_HTTPS" json:"enable_https"`
}

func (c *ClientConfig) populateEmptyFields(another ClientConfig) {
	if c.BaseDir == "" && another.BaseDir != "" {
		c.BaseDir = another.BaseDir
	}
	if c.SyncServerURL == "" && another.SyncServerURL != "" {
		c.SyncServerURL = another.SyncServerURL
	}
	if c.SyncPeriod == 30 && another.SyncPeriod != 30 {
		c.SyncPeriod = another.SyncPeriod
	}
	if !c.HTTPSEnabled && another.HTTPSEnabled {
		c.HTTPSEnabled = another.HTTPSEnabled
	}
}

// LoadClientConfig reads environment variables and flags, prior to flags.
func LoadClientConfig() (ClientConfig, error) {

	var mainConfig ClientConfig

	flag.StringVar(&mainConfig.BaseDir, "baseDir", "", "base directory where Gophkeeper will store data")
	flag.StringVar(&mainConfig.SyncServerURL, "server", "", "gophkeeper synchronization server")
	flag.Int64Var(&mainConfig.SyncPeriod, "syncPeriod", 30, "gophkeeper synchronization period, in seconds")
	flag.BoolVar(&mainConfig.HTTPSEnabled, "t", true, "enable HTTPS with self signed certificate")

	flag.Parse()

	var envsConfig ClientConfig
	if err := env.Parse(&envsConfig); err != nil {
		panic(err)
	}

	mainConfig.populateEmptyFields(envsConfig)

	if mainConfig.BaseDir == "" {
		dir, err := os.UserHomeDir()
		if err != nil {
			return ClientConfig{}, err
		}
		mainConfig.BaseDir = dir
	}
	mainConfig.BaseDir = filepath.Join(mainConfig.BaseDir, ".gophkeeper")
	err := os.MkdirAll(mainConfig.BaseDir, 0755)
	if err != nil {
		panic(err)
	}

	return mainConfig, nil
}
