package server

// Player is a human player connected to the game server.
type Player struct {
	Username string
	Password string
	Handler  *ClientHandler
}

// NewPlayer returns a new player model.
func NewPlayer(username, password string) *Player {
	return &Player{
		Username: username,
		Password: password,
	}
}
