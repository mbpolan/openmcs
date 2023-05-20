package game

// ScriptHandler is the interface between the game engine and the script system.
type ScriptHandler interface {
	// handleRemovePlayer schedules a player to be removed from the game.
	handleRemovePlayer(pe *playerEntity)
}
