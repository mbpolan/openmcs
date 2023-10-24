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
	// handleSetInterfaceColor sends a player's client a color to set for an interface.
	handleSetInterfaceColor(pe *playerEntity, interfaceID int, color model.Color)
	// handleSetInterfaceModel sends a player's client an item model to show on an interface.
	handleSetInterfaceModel(pe *playerEntity, interfaceID, itemID, zoom int)
	// handleSetInterfaceText sends a player's client text to show on an interface.
	handleSetInterfaceText(pe *playerEntity, interfaceID int, text string)
	// handleSetInterfaceSetting sends a setting value for the current interface.
	handleSetInterfaceSetting(pe *playerEntity, settingID, value int)
	// handleRemovePlayer schedules a player to be removed from the game.
	handleRemovePlayer(pe *playerEntity)
	// handleConsumeInventoryItems attempts to consume a set of items from the player's inventory, returning true if successful
	// or false if not. itemIDsAmounts should be a vararg slice consisting of the item ID followed by the amount.
	handleConsumeInventoryItems(pe *playerEntity, itemIDsAmounts ...int) bool
	// handleConsumeInventoryItemInSlot attempts to consume an item at a particular slot in the player's inventory,
	// returning true if successful or false if not.
	handleConsumeInventoryItemInSlot(pe *playerEntity, slotID, amount int) bool
	// handleAddInventoryItem adds an item with an amount to the player's inventory. If the player's inventory is full,
	// the item is dropped on the ground instead.
	handleAddInventoryItem(pe *playerEntity, itemID, amount int)
	// handleCountInventoryItems returns the number of items that existing in the player's inventory. If an item is
	// stackable, the total number of stacked items of that kind will be returned. Otherwise, a count of each instance
	// of a non-stackable item will be returned,
	handleCountInventoryItems(pe *playerEntity, itemID int) int
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
	// handleGrantExperience grants a player experience points in a skill.
	handleGrantExperience(pe *playerEntity, skillType model.SkillType, experience float64)
	// handleSetSidebarTab sets the active tab on the client's sidebar.
	handleSetSidebarTab(pe *playerEntity, tab model.ClientTab)
	// handleChangePlayerMovementSpeed changes the movement speed of a player.
	handleChangePlayerMovementSpeed(pe *playerEntity, speed model.MovementSpeed)
	// handleChangePlayerAutoRetaliate changes a player's auto-retaliate combat option.
	handleChangePlayerAutoRetaliate(pe *playerEntity, enabled bool)
	// handleSetPlayerQuestStatus updates the status of a quest for a player.
	handleSetPlayerQuestStatus(pe *playerEntity, questID int, status model.QuestStatus)
	// handleSetPlayerQuestFlag sets a quest flag with a value for a player.
	handleSetPlayerQuestFlag(pe *playerEntity, questID, flagID, value int)
	// handleSetPlayerMusicTrackUnlocked sets a music track as (un)locked for a player.
	handleSetPlayerMusicTrackUnlocked(pe *playerEntity, musicID int, enabled bool)
	// handlePlayMusic sends the player's client a music track to play.
	handlePlayMusic(pe *playerEntity, musicID int)
	// handleShowInterface shows an interface for a player.
	handleShowInterface(pe *playerEntity, interfaceID int)
	// handleDelayCurrentAction blocks the player from performing other actions until a set amount of game ticks have
	// elapsed.
	handleDelayCurrentAction(pe *playerEntity, tickDuration int)
	// handleActivatePrayer enables a prayer.
	handleActivatePrayer(pe *playerEntity, prayerID, drain int)
	// handleDeactivatePrayer disables a prayer.
	handleDeactivatePrayer(pe *playerEntity, prayerID int)
}
