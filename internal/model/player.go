package model

type PlayerType int

const (
	PlayerNormal PlayerType = iota
	PlayerModerator
	PlayerAdmin
)

// Player is a human player connected to the game server.
type Player struct {
	ID         int
	Username   string
	Password   string
	Type       PlayerType
	Flagged    bool
	GlobalPos  Vector3D
	Appearance *EntityAppearance
}

// NewPlayer returns a new player model.
func NewPlayer(id int, username, password string, pType PlayerType, flagged bool, globalPos Vector3D) *Player {
	// define a default appearance
	appearance := &EntityAppearance{
		Equipment: [12]int{
			256,  // head (256 - 265)
			266,  // beard (266 - 273)
			274,  // torso (274 - 281)
			282,  // arms (282 - 288)
			292,  // legs (292 - 297)
			298,  // boots (298 - ???)
			289,  // hands (289 - 291)
			1564, // worn item (l cape)
			1552, // head accessory (y phat)
			1699, // shield (d sq)
			1817, // weapon (d long)
			2216, // necklace (aog)
		},
		Body: [5]int{
			0,
			0,
			0,
			0,
			0,
		},
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
		ID:         id,
		Username:   username,
		Password:   password,
		Type:       pType,
		Flagged:    flagged,
		GlobalPos:  globalPos,
		Appearance: appearance,
	}
}
