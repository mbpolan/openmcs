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

// InterfaceCondition is a condition for an interface to be active.
type InterfaceCondition struct {
	Type  int
	Value int
}

// Interface is a client-side interface that displays information and allows actions.
type Interface struct {
	// ID is the identifier for this interface.
	ID int
	// Parent is the parent interface.
	Parent *Interface
	// Children is a slice of child interfaces.
	Children []*Interface
	// Actions is a slice of actions that this interface supports.
	Actions []string
	// Conditions is a slice of conditions for this interface.
	Conditions []InterfaceCondition
	// OpCodes is a map of op codes to their sub codes
	OpCodes map[int][]int
}
