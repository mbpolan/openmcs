package model

// FriendsListStatus enumerates which state a player's client should show the friends list.
type FriendsListStatus int

const (
	FriendsListStatusLoading FriendsListStatus = iota
	FriendsListStatusPending
	FriendsListStatusLoaded
)

// InteractionMode controls from whom a player receives trades and duels requests.
type InteractionMode int

const (
	InteractionModePublic InteractionMode = iota
	InteractionModeFriends
	InteractionModeOff
)

// ClientTab enumerates the possible client sidebar tabs available for displaying interface.
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

// ClientTabs is a slice of all client tab interface.
var ClientTabs = []ClientTab{
	ClientTabEquippedItem,
	ClientTabSkills,
	ClientTabQuests,
	ClientTabInventory,
	ClientTabEquipment,
	ClientTabPrayers,
	ClientTabSpells,
	ClientTabFriendsList,
	ClientTabIgnoreList,
	ClientTabLogout,
	ClientTabSettings,
	ClientTabControls,
	ClientTabMusic,
}
