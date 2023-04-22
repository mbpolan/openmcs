package driver

import (
	"github.com/golang-migrate/migrate/v4/database"
	"github.com/mbpolan/openmcs/internal/model"
)

// Driver is an interface for a driver that interfaces with a backend database.
type Driver interface {
	// Migration returns a handle to the underlying store for use with migrations.
	Migration() (database.Driver, error)

	// LoadPlayer loads data about a player with a username.
	LoadPlayer(username string) (*model.Player, error)

	// Close cleans up resources used by the driver.
	Close() error
}