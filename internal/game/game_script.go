package game

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
	// handleSendServerMessage sends a server message to a player.
	handleSendServerMessage(pe *playerEntity, message string)
}
