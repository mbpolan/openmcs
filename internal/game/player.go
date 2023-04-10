package game

type PlayerType int

const (
	PlayerNormal PlayerType = iota
	PlayerModerator
	PlayerAdmin
)

// Player is a human player connected to the game server.
type Player struct {
	Username string
	Password string
	Type     PlayerType
	Flagged  bool
}

// NewPlayer returns a new player model.
func NewPlayer(username, password string, pType PlayerType, flagged bool) *Player {
	return &Player{
		Username: username,
		Password: password,
		Type:     pType,
		Flagged:  flagged,
	}
}
