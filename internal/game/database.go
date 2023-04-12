package game

import "github.com/mbpolan/openmcs/internal/model"

// Database provides access to the server's persistent data storage.
type Database struct {
}

// NewDatabase creates a new database provider.
func NewDatabase() *Database {
	return &Database{}
}

// LoadPlayer returns a player's data based on their username. If the player does not exist, nil is returned.
func (d *Database) LoadPlayer(username string) (*model.Player, error) {
	// TODO: use a real database
	if username != "mike" {
		return nil, nil
	}

	return model.NewPlayer(username, "foo", model.PlayerNormal, false), nil
}
