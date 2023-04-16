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
