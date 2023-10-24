package game

import "github.com/mbpolan/openmcs/internal/model"

// ActionPriority enumerates the priority of an action when it's deferred.
type ActionPriority int

const (
	// ActionPriorityNormal is an action scheduled after all currently deferred actions.
	ActionPriorityNormal ActionPriority = iota
	// ActionPriorityHigh is an action scheduled ahead of all currently deferred actions.
	ActionPriorityHigh
)

// ActionType enumerates deferred actions that a player can take.
type ActionType int

const (
	ActionMoveInventoryItem ActionType = iota
	ActionSendServerMessage
	ActionSendSkills
	ActionSendModes
	ActionSendFriendList
	ActionSendIgnoreList
	ActionSendEquipment
	ActionSendInventory
	ActionTakeGroundItem
	ActionDropInventoryItem
	ActionEquipItem
	ActionUnequipItem
	ActionShowInterface
	ActionHideInterfaces
	ActionDoInterfaceAction
	ActionTeleportPlayer
	ActionCastSpellOnItem
	ActionExperienceGrant
	ActionSendRunEnergy
	ActionSendChangeEvent
	ActionSendWeight
	ActionPlayMusic
	ActionDelayCurrent
)

// Action is an action that will be performed after a number of game ticks have elapsed.
type Action struct {
	ActionType              ActionType
	TickDelay               uint
	ServerMessageAction     *ServerMessageAction
	MoveInventoryItemAction *MoveInventoryItemAction
	TakeGroundItem          *TakeGroundItemAction
	DropInventoryItemAction *DropInventoryItemAction
	EquipItemAction         *EquipItemAction
	UnequipItemAction       *UnequipItemAction
	ShowInterfaceAction     *ShowInterfaceAction
	DoInterfaceAction       *DoInterfaceAction
	TeleportPlayerAction    *TeleportPlayerAction
	CastSpellOnItemAction   *CastSpellOnItemAction
	ExperienceGrantAction   *ExperienceGrantAction
	PlayMusicAction         *PlayMusicAction
	PlayerChangeEventAction *PlayerChangeEventAction
	SendSkillsAction        *SendSkillsAction
}

// ActionResult is a bitmask describing the mutations resulting from an action.
type ActionResult int

const (
	ActionResultNoChange ActionResult = 0
	ActionResultPending               = 1 << iota
	ActionResultChangeRegions
	ActionResultClearAnimations
	ActionResultClearGraphics
)

// ServerMessageAction is an action to send the player a server message.
type ServerMessageAction struct {
	Message string
}

// MoveInventoryItemAction is an action to move or swap the position of an inventory item.
type MoveInventoryItemAction struct {
	FromSlot int
	ToSlot   int
}

// TakeGroundItemAction is an action to pick up a ground item that should occur at a position.
type TakeGroundItemAction struct {
	GlobalPos model.Vector3D
	Item      *model.Item
}

// DropInventoryItemAction is an action to drop an inventory item.
type DropInventoryItemAction struct {
	InterfaceID       int
	Item              *model.Item
	SecondaryActionID int
}

// EquipItemAction is an action to equip an item from the player's inventory
type EquipItemAction struct {
	InterfaceID int
	Item        *model.Item
}

// UnequipItemAction is an action to unequip an item from the player's equipment
type UnequipItemAction struct {
	InterfaceID int
	Item        *model.Item
	SlotType    model.EquipmentSlotType
}

// ShowInterfaceAction is an action to show an interface.
type ShowInterfaceAction struct {
	InterfaceID int
}

// DoInterfaceAction is an action taken on an interface.
type DoInterfaceAction struct {
	Parent *model.Interface
	Actor  *model.Interface
}

// TeleportPlayerAction is an action to teleport a player.
type TeleportPlayerAction struct {
	GlobalPos model.Vector3D
}

// CastSpellOnItemAction is an action to cast a spell on an inventory item.
type CastSpellOnItemAction struct {
	SlotID               int
	ItemID               int
	InventoryInterfaceID int
	SpellInterfaceID     int
}

// ExperienceGrantAction is an action to grant the player experience after a delay.
type ExperienceGrantAction struct {
	SkillType  model.SkillType
	Experience float64
}

// PlayMusicAction is an action to instruct the player's client to play a song.
type PlayMusicAction struct {
	MusicID int
}

// PlayerChangeEventAction is an action to handle a change to a player's attributes.
type PlayerChangeEventAction struct {
	Event model.PlayerChangeEvent
}

// SendSkillsAction is an action that sends skill information to the player's client.
type SendSkillsAction struct {
	SkillTypes []model.SkillType
}
