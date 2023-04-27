package game

import (
	"strconv"
	"strings"
)

// ChatCommandType enumerates possible chat commands recognized by the server.
type ChatCommandType int

const (
	ChatCommandTypeSpawnItem ChatCommandType = iota
	ChatCommandTypeClearTile
)

// ChatCommandSpawnItemParams contains parameters for a chat command that spawns a ground item.
type ChatCommandSpawnItemParams struct {
	ItemID int
}

// ChatCommand is a game command embedded in a player chat message.
type ChatCommand struct {
	Type      ChatCommandType
	SpawnItem *ChatCommandSpawnItemParams
}

// ParseChatCommand attempts to parse a chat command from a string of text. If no recognized command is found, then
// nil is returned instead.
func ParseChatCommand(text string) *ChatCommand {
	parts := strings.Split(strings.ToLower(text), " ")
	if len(parts) == 0 {
		return nil
	}

	// the command is the first element and optional arguments follow
	command := parts[0]
	args := parts[1:]

	switch command {
	case "i":
		// spawn a ground item
		if len(args) != 1 {
			return nil
		}

		// only argument is a numeric item id
		itemID, err := strconv.Atoi(args[0])
		if err != nil {
			return nil
		}

		return &ChatCommand{
			Type:      ChatCommandTypeSpawnItem,
			SpawnItem: &ChatCommandSpawnItemParams{ItemID: itemID},
		}

	case "ct":
		return &ChatCommand{
			Type: ChatCommandTypeClearTile,
		}

	default:
	}

	return nil
}
