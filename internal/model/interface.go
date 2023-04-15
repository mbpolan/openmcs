package model

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
