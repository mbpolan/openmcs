package model

// InteractionMode controls from whom a player receives trades and duels requests.
type InteractionMode int

const (
	InteractionModePublic InteractionMode = iota
	InteractionModeFriends
	InteractionModeOff
)

// ClientTab enumerates the possible client sidebar tabs available for displaying interfaces.
type ClientTab int

const (
	ClientTabEquippedItem ClientTab = iota
	ClientTabSkills
	ClientTabQuests
	ClientTabInventory
	ClientTabEquipment
	ClientTabPrayers
	ClientTabSpells
	ClientTabFriendsList
	ClientTabIgnoreList
	ClientTabLogout
	ClientTabSettings
	ClientTabControls
	ClientTabMusic
)
