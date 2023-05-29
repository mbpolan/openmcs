package game

import (
	"github.com/mbpolan/openmcs/internal/model"
)

// ScriptHandler is the interface between the game engine and the script system.
type ScriptHandler interface {
	// handleSetSidebarInterface sends a player's client an interface to show on a sidebar tab.
	handleSetSidebarInterface(pe *playerEntity, interfaceID, sidebarID int)
	// handleClearSidebarInterface sends a player's client a command to remove an interface on a sidebar tab.
	handleClearSidebarInterface(pe *playerEntity, sidebarID int)
	// handleSetInterfaceModel sends a player's client an item model to show on an interface.
	handleSetInterfaceModel(pe *playerEntity, interfaceID, itemID, zoom int)
	// handleSetInterfaceText sends a player's client text to show on an interface.
	handleSetInterfaceText(pe *playerEntity, interfaceID int, text string)
	// handleSetInterfaceSetting sends a setting value for the current interface.
	handleSetInterfaceSetting(pe *playerEntity, settingID, value int)
	// handleRemovePlayer schedules a player to be removed from the game.
	handleRemovePlayer(pe *playerEntity)
	// handleConsumeRunes attempts to consume a set of runes from the player's inventory, returning true if successful or
	// false if not. runeIDsAmounts should be a vararg slice consisting of the rune item ID followed by the amount.
	handleConsumeRunes(pe *playerEntity, runeIDsAmounts ...int) bool
	// handleConsumeInventoryItem removes an inventory item, or decrements its amount by one, returning true if
	// successful.
	handleConsumeInventoryItem(pe *playerEntity, slotID int) bool
	// handleAddInventoryItem adds an item with an amount to the player's inventory. If the player's inventory is full,
	// the item is dropped on the ground instead.
	handleAddInventoryItem(pe *playerEntity, itemID, amount int)
	// handleSendServerMessage sends a server message to a player.
	handleSendServerMessage(pe *playerEntity, message string)
	// handleTeleportPlayer teleports a player to another location.
	handleTeleportPlayer(pe *playerEntity, globalPos model.Vector3D)
	// handleAnimatePlayer sets a player's current animation with an expiration after a number of game ticks.
	handleAnimatePlayer(pe *playerEntity, animationID, tickDuration int)
	// handleSetPlayerGraphic sets a graphic to display with the player model at a height offset from the ground. A
	// client-side tick delay can be provided to delay the start of the graphic being applied, and an expiration after a
	// number of game ticks when the graphic will be removed.
	handleSetPlayerGraphic(pe *playerEntity, graphicID, height, delay, tickDuration int)
	// handleGrantExperience grants a player experience, delaying the current action for an amount of game ticks.
	handleGrantExperience(pe *playerEntity, skillType model.SkillType, experience, tickDelay int)
	// handleSetSidebarTab sets the active tab on the client's sidebar.
	handleSetSidebarTab(pe *playerEntity, tab model.ClientTab)
}
