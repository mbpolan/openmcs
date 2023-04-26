package game

import (
	"strconv"
	"strings"
)

// ChatCommandType enumerates possible chat commands recognized by the server.
type ChatCommandType int

const (
	ChatCommandTypeSpawnItem ChatCommandType = iota
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

// ParseChatCommand attempts to parse a chat command from a string of text. If no command is found, nil is returned.
func ParseChatCommand(text string) *ChatCommand {
	if !strings.HasPrefix(text, "!") {
		return nil
	}

	parts := strings.Split(strings.ToLower(text), " ")
	if len(parts) == 0 {
		return nil
	}

	command := strings.TrimPrefix(parts[0], "!")
	args := parts[1:]

	switch command {
	case "i":
		// spawn a ground item
		if len(args) != 1 {
			return nil
		}

		itemID, err := strconv.Atoi(args[0])
		if err != nil {
			return nil
		}

		return &ChatCommand{
			Type:      ChatCommandTypeSpawnItem,
			SpawnItem: &ChatCommandSpawnItemParams{ItemID: itemID},
		}
	default:
	}

	return nil
}
