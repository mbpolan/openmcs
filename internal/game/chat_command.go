package game

import (
	"github.com/mbpolan/openmcs/internal/model"
	"strconv"
	"strings"
)

// ChatCommandType enumerates possible chat commands recognized by the server.
type ChatCommandType int

const (
	ChatCommandTypeSpawnItem ChatCommandType = iota
	ChatCommandTypeClearTile
	ChatCommandTypePosition
	ChatCommandTeleport
	ChatCommandTeleportRelative
	ChatCommandShowInterface
	ChatCommandHideInterfaces
	ChatCommandCharacterDesigner
	ChatCommandReloadScripts
)

// ChatCommandSpawnItemParams contains parameters for a chat command that spawns a ground Item.
type ChatCommandSpawnItemParams struct {
	ItemID             int
	Amount             int
	DespawnTimeSeconds *int
}

// ChatCommandShowInterfaceParams contains parameters for showing an interface.
type ChatCommandShowInterfaceParams struct {
	InterfaceID int
}

// ChatCommand is a game command embedded in a player chat message.
type ChatCommand struct {
	Type          ChatCommandType
	Pos           model.Vector3D
	SpawnItem     *ChatCommandSpawnItemParams
	ShowInterface *ChatCommandShowInterfaceParams
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
		// spawn a ground Item
		if len(args) < 1 {
			return nil
		}

		// first required argument is a numeric Item id
		itemID, err := strconv.Atoi(args[0])
		if err != nil {
			return nil
		}

		// second optional argument is the stack amount
		amount := 1
		if len(args) > 1 {
			i, err := strconv.Atoi(args[1])
			if err != nil {
				return nil
			}

			amount = i
		}

		// third optional argument is a despawn time in seconds
		var despawnTimeSeconds *int
		if len(args) > 2 {
			timeout, err := strconv.Atoi(args[1])
			if err != nil {
				return nil
			}

			despawnTimeSeconds = &timeout
		}

		return &ChatCommand{
			Type: ChatCommandTypeSpawnItem,
			SpawnItem: &ChatCommandSpawnItemParams{
				ItemID:             itemID,
				Amount:             amount,
				DespawnTimeSeconds: despawnTimeSeconds,
			},
		}

	case "ct":
		return &ChatCommand{
			Type: ChatCommandTypeClearTile,
		}

	case "tpr":
		if len(args) == 0 {
			return nil
		}

		// teleport to a location relative to the player's current position
		dx, dy, dz := 0, 0, 0
		var err error

		// arguments are the x-, y- and z-coordinates, in that order, with each coordinate being optional
		if len(args) > 0 {
			dx, err = strconv.Atoi(args[0])
			if err != nil {
				return nil
			}
		}

		if len(args) > 1 {
			dy, err = strconv.Atoi(args[1])
			if err != nil {
				return nil
			}
		}

		if len(args) > 2 {
			dz, err = strconv.Atoi(args[2])
			if err != nil {
				return nil
			}
		}

		return &ChatCommand{
			Type: ChatCommandTeleportRelative,
			Pos: model.Vector3D{
				X: dx,
				Y: dy,
				Z: dz,
			},
		}

	case "tp":
		// teleport to a location
		if len(args) != 3 {
			return nil
		}

		// arguments are the x-, y- and z-coordinates, in that order
		x, err := strconv.Atoi(args[0])
		if err != nil {
			return nil
		}

		y, err := strconv.Atoi(args[1])
		if err != nil {
			return nil
		}

		z, err := strconv.Atoi(args[2])
		if err != nil {
			return nil
		}

		return &ChatCommand{
			Type: ChatCommandTeleport,
			Pos: model.Vector3D{
				X: x,
				Y: y,
				Z: z,
			},
		}

	case "pos":
		return &ChatCommand{
			Type: ChatCommandTypePosition,
		}

	case "design":
		return &ChatCommand{
			Type: ChatCommandCharacterDesigner,
		}

	case "inf":
		// show an interface
		if len(args) < 1 {
			return nil
		}

		// first required argument is a numeric interface id
		interfaceID, err := strconv.Atoi(args[0])
		if err != nil {
			return nil
		}

		return &ChatCommand{
			Type: ChatCommandShowInterface,
			ShowInterface: &ChatCommandShowInterfaceParams{
				InterfaceID: interfaceID,
			},
		}

	case "clear":
		// clear open interfaces
		return &ChatCommand{
			Type: ChatCommandHideInterfaces,
		}

	case "rscripts":
		// reload script manager
		return &ChatCommand{
			Type: ChatCommandReloadScripts,
		}

	default:
	}

	return nil
}
