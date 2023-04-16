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
	if username != "mike" && username != "hurz" {
		return nil, nil
	}

	// TODO: maintain player position
	globalPos := model.Vector3D{
		X: 3116,
		Y: 3116,
		Z: 0,
	}

	return model.NewPlayer(int(username[0]), username, "foo", model.PlayerNormal, false, globalPos), nil
}
