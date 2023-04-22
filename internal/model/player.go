package model

import "strings"

type PlayerType int

const (
	PlayerNormal PlayerType = iota
	PlayerModerator
	PlayerAdmin
)

// Player is a human player connected to the game server. This struct stores a player's persistent data, including
// various preferences, game world properties and other such attributes.
type Player struct {
	ID           int
	Username     string
	PasswordHash string
	Type         PlayerType
	Flagged      bool
	GlobalPos    Vector3D
	Appearance   *EntityAppearance
	Modes        PlayerModes
	Muted        bool
	Friends      []string
	Ignored      []string
}

// PlayerModes indicates what types of chat and interactions a player wishes to receive.
type PlayerModes struct {
	PublicChat  ChatMode
	PrivateChat ChatMode
	Interaction InteractionMode
}

// NewPlayer returns a new player model.
func NewPlayer(username string) *Player {
	// define a default appearance
	appearance := &EntityAppearance{
		Equipment: [12]int{},
		Body:      [5]int{},
		Animations: map[AnimationID]int{
			AnimationStand:     0x080D, // standing
			AnimationStandTurn: 0xFFFF, // turning
			AnimationWalk:      0x067C, // walk
			AnimationTurnAbout: 0xFFFF, // turn about
			AnimationTurnRight: 0xFFFF, // turn right
			AnimationTurnLeft:  0xFFFF, // turn left
			AnimationRun:       0x067D, // run
		},
		Gender:         EntityMale,
		OverheadIconID: 0,
		CombatLevel:    3,
		SkillLevel:     200,
		Updated:        false,
	}

	return &Player{
		Username:   username,
		Appearance: appearance,
	}
}

// HasFriend determines if the given player username is on this player's friends list.
func (p *Player) HasFriend(username string) bool {
	target := strings.ToLower(username)
	for _, friend := range p.Friends {
		if strings.ToLower(friend) == target {
			return true
		}
	}

	return false
}

// IsIgnored determines if the given player username is on this player's ignore list.
func (p *Player) IsIgnored(username string) bool {
	target := strings.ToLower(username)
	for _, ignored := range p.Ignored {
		if strings.ToLower(ignored) == target {
			return true
		}
	}

	return false
}
