package db

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/apolsh/yapr-gophkeeper/internal/logger"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/source"
	// migrate tools
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

const _defaultAttempts = 5
const _defaultTimeout = 10 * time.Second

var log = logger.LoggerOfComponent("migrate")

// RunMigration starts the database schema migration.
func RunMigration(migrationSourceDriver source.Driver, databaseURL string) error {

	if !strings.Contains(databaseURL, "sslmode") {
		databaseURL += "?sslmode=disable"
	}

	var (
		attempts = _defaultAttempts
		err      error
		m        *migrate.Migrate
	)

	for attempts > 0 {
		//m, err = migrate.New("file://migrations", databaseURL)
		m, err = migrate.NewWithSourceInstance("iofs", migrationSourceDriver, databaseURL)

		if err == nil {
			log.Error(err)
			break
		}

		log.Info("database is trying to connect, attempts left: %d", attempts)
		time.Sleep(_defaultTimeout)
		attempts--
	}

	if err != nil {
		err = fmt.Errorf("database connect error: %s", err)
		log.Error(err)
		return err
	}

	if m == nil {
		err = errors.New("failed to initialise migration instance")
		log.Error(err)
		return err
	}

	err = m.Up()
	defer func(m *migrate.Migrate) {
		err, _ := m.Close()
		if err != nil {
			log.Error(err)
		}
	}(m)
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		err = fmt.Errorf("up error: %s", err)
		log.Error(err)
		return err
	}

	if errors.Is(err, migrate.ErrNoChange) {
		log.Info("no change")
		return nil
	}

	log.Info("up success")
	return nil
}
