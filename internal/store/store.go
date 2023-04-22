package store

import (
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/mbpolan/openmcs/internal/config"
	"github.com/mbpolan/openmcs/internal/logger"
	"github.com/mbpolan/openmcs/internal/model"
	"github.com/mbpolan/openmcs/internal/store/driver"
	"github.com/pkg/errors"
	"strings"
)

// Store is a backend database used for persistent storage of game data.
type Store struct {
	config *config.Config
	driver driver.Driver
}

// New creates a new persistent storage provider.
func New(cfg *config.Config) (*Store, error) {
	var dbDriver driver.Driver
	var err error

	// initialize a backend driver based on configuration
	switch strings.ToLower(cfg.Store.Driver) {
	case "sqlite3":
		if cfg.Store.SQLite3 == nil {
			return nil, fmt.Errorf("missing database.sqlite3 configuration")
		}

		dbDriver, err = driver.NewSQLite3Driver(cfg.Store.SQLite3)
	default:
		err = fmt.Errorf("unsupported driver: %s", cfg.Store.Driver)
	}

	if err != nil {
		return nil, errors.Wrapf(err, "failed to initialize %s driver", cfg.Store.Driver)
	}

	return &Store{
		config: cfg,
		driver: dbDriver,
	}, nil
}

// Migrate runs migrations against the backend persistent store.
func (s *Store) Migrate() error {
	logger.Debugf("running migrations from %s", s.config.Store.MigrationsDir)
	sourceDir := fmt.Sprintf("file://%s", s.config.Store.MigrationsDir)

	// get a handle to the driver for migrations
	handle, err := s.driver.Migration()
	if err != nil {
		return errors.Wrap(err, "failed to get handle to migration driver")
	}

	// use the same driver as the persistent store uses
	m, err := migrate.NewWithDatabaseInstance(sourceDir, s.config.Store.Driver, handle)
	if err != nil {
		return err
	}

	// run migrations, ignoring "error" reported when there are no detected changes
	err = m.Up()
	if err != migrate.ErrNoChange {
		return err
	}

	logger.Infof("migrations successfully completed")

	return nil
}

// Close cleans up resources used by the persistent store provider.
func (s *Store) Close() error {
	return s.driver.Close()
}

// LoadPlayer loads information about a player.
func (s *Store) LoadPlayer(username string) (*model.Player, error) {
	return s.driver.LoadPlayer(username)
}
