package driver

import (
	"database/sql"
	"github.com/mbpolan/openmcs/internal/config"
	"github.com/mbpolan/openmcs/internal/model"
	_ "modernc.org/sqlite"
)

// SQLite3Driver is a driver that interfaces with a SQLite3 database.
type SQLite3Driver struct {
	db *sql.DB
}

// NewSQLite3Driver creates a new SQLite3 database driver.
func NewSQLite3Driver(cfg *config.SQLite3DatabaseConfig) (Driver, error) {
	db, err := sql.Open("sqlite", cfg.URI)
	if err != nil {
		return nil, err
	}

	return &SQLite3Driver{
		db: db,
	}, nil
}

// LoadPlayer loads information about a player from a SQLite3 database.
func (s *SQLite3Driver) LoadPlayer(username string) (*model.Player, error) {
	// TODO: use a real database
	if username != "mike" && username != "hurz" {
		return nil, nil
	}

	// TODO: just for testing
	x := 0
	y := 0
	if username == "hurz" {
		x = 2
		y = -2
	}

	// TODO: maintain player position
	globalPos := model.Vector3D{
		X: 3116 + x,
		Y: 3116 + y,
		Z: 0,
	}

	return model.NewPlayer(int(username[0]), username, "", model.PlayerNormal, false, globalPos), nil
}

// Close cleans up resources used by the SQLite3 driver.
func (s *SQLite3Driver) Close() error {
	return s.db.Close()
}
