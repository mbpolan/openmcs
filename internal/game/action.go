package game

import "github.com/mbpolan/openmcs/internal/model"

// ActionType enumerates deferred actions that a player can take.
type ActionType int

const (
	ActionMoveInventoryItem ActionType = iota
	ActionSendServerMessage
	ActionSendSkills
	ActionSendInterfaces
	ActionSendFriendList
	ActionSendIgnoreList
	ActionSendEquipment
	ActionSendInventory
	ActionTakeGroundItem
	ActionDropInventoryItem
	ActionEquipItem
	ActionUnequipItem
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
}

// ServerMessageAction is an action to send the player a server message.
type ServerMessageAction struct {
	Message string
}

// MoveInventoryItemAction is an action to move or swap the position of an inventory item.
type MoveInventoryItemAction struct {
	FromSlot int
	ToSlot   int
}

// TakeGroundItemAction is an action to pick up a ground Item that should occur at a position.
type TakeGroundItemAction struct {
	GlobalPos model.Vector3D
	Item      *model.Item
}

// DropInventoryItemAction is an action to drop an inventory Item.
type DropInventoryItemAction struct {
	InterfaceID       int
	Item              *model.Item
	SecondaryActionID int
}

// EquipItemAction is an action to equip an Item from the player's inventory
type EquipItemAction struct {
	InterfaceID int
	Item        *model.Item
}

// UnequipItemAction is an action to unequip an Item from the player's equipment
type UnequipItemAction struct {
	InterfaceID int
	Item        *model.Item
	SlotType    model.EquipmentSlotType
}
