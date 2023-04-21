package store

import (
	"fmt"
	"github.com/mbpolan/openmcs/internal/config"
	"github.com/mbpolan/openmcs/internal/model"
	"github.com/mbpolan/openmcs/internal/store/driver"
	"github.com/pkg/errors"
	"strings"
)

// Store is a backend database used for persistent storage of game data.
type Store struct {
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
		driver: dbDriver,
	}, nil
}

// Close cleans up resources used by the persistent store provider.
func (s *Store) Close() error {
	return s.driver.Close()
}

// LoadPlayer loads information about a player.
func (s *Store) LoadPlayer(username string) (*model.Player, error) {
	return s.driver.LoadPlayer(username)
}
