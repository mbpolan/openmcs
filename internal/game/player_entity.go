package game

import (
	"github.com/mbpolan/openmcs/internal/model"
	"github.com/mbpolan/openmcs/internal/network"
	"github.com/mbpolan/openmcs/internal/network/response"
	"sync"
	"time"
)

// pendingActionType enumerates deferred actions that a player can take.
type pendingActionType int

const (
	pendingActionTakeGroundItem pendingActionType = iota
	pendingActionDropInventoryItem
	pendingActionEquipItem
	pendingActionUnequipItem
)

// pendingAction is a deferred action that a player has requested be done.
type pendingAction struct {
	actionType              pendingActionType
	takeGroundItem          *takeGroundItemAction
	dropInventoryItemAction *dropInventoryItemAction
	equipItemAction         *equipItemAction
	unequipItemAction       *unequipItemAction
}

// takeGroundItemAction is an action to pick up a ground item that should occur at a position.
type takeGroundItemAction struct {
	globalPos model.Vector3D
	item      *model.Item
}

// dropInventoryItemAction is an action to drop an inventory item.
type dropInventoryItemAction struct {
	interfaceID       int
	item              *model.Item
	secondaryActionID int
}

// equipItemAction is an action to equip an item from the player's inventory
type equipItemAction struct {
	interfaceID int
	item        *model.Item
}

// unequipItemAction is an action to unequip an item from the player's equipment
type unequipItemAction struct {
	interfaceID int
	item        *model.Item
	slotType    model.EquipmentSlotType
}

// playerEntity represents a player and their state while they are logged into the game world.
type playerEntity struct {
	lastInteraction     time.Time
	player              *model.Player
	tracking            map[int]*playerEntity
	changeChan          chan bool
	doneChan            chan bool
	updateChan          chan *response.PlayerUpdateResponse
	path                []model.Vector2D
	nextPathIdx         int
	scheduler           *Scheduler
	writer              *network.ProtocolWriter
	lastWalkTime        time.Time
	lastChatMessage     *model.ChatMessage
	lastChatTime        time.Time
	chatHighWater       time.Time
	tabInterfaces       map[model.ClientTab]int
	teleportGlobal      *model.Vector3D
	privateMessageID    int
	regionOrigin        model.Vector2D
	nextStatusBroadcast *playerStatusBroadcast
	nextUpdate          *response.PlayerUpdateResponse
	deferredAction      *pendingAction
	mu                  sync.Mutex
}

type playerStatusBroadcast struct {
	targets []string
}

// newPlayerEntity creates a new player entity.
func newPlayerEntity(p *model.Player, w *network.ProtocolWriter) *playerEntity {
	changeChan := make(chan bool)

	return &playerEntity{
		lastInteraction:  time.Now(),
		player:           p,
		tracking:         map[int]*playerEntity{},
		changeChan:       changeChan,
		doneChan:         make(chan bool, 1),
		updateChan:       make(chan *response.PlayerUpdateResponse, 1),
		scheduler:        NewScheduler(changeChan),
		privateMessageID: 1,
		writer:           w,
	}
}

// MarkStatusBroadcast marks that this player's online/offline status should be broadcast to everyone.
func (pe *playerEntity) MarkStatusBroadcast() {
	if pe.nextStatusBroadcast == nil {
		pe.nextStatusBroadcast = &playerStatusBroadcast{}
	}

	pe.nextStatusBroadcast.targets = nil
}

// MarkStatusBroadcastTarget adds a target to receive this player's online/offline status.
func (pe *playerEntity) MarkStatusBroadcastTarget(target string) {
	if pe.nextStatusBroadcast == nil {
		pe.nextStatusBroadcast = &playerStatusBroadcast{}
	}

	pe.nextStatusBroadcast.targets = append(pe.nextStatusBroadcast.targets, target)
}

// DeferTakeGroundItemAction sets the player's pending action to pick up a specific ground item at a position, in
// global coordinates. This will overwrite any previously deferred action.
func (pe *playerEntity) DeferTakeGroundItemAction(item *model.Item, globalPos model.Vector3D) {
	pe.deferredAction = &pendingAction{
		actionType: pendingActionTakeGroundItem,
		takeGroundItem: &takeGroundItemAction{
			globalPos: globalPos,
			item:      item,
		},
	}
}

// DeferDropInventoryItem sets the player's pending action to drop an inventory item. This will overwrite any previously
// deferred action.
func (pe *playerEntity) DeferDropInventoryItem(item *model.Item, interfaceID, secondaryActionID int) {
	pe.deferredAction = &pendingAction{
		actionType: pendingActionDropInventoryItem,
		dropInventoryItemAction: &dropInventoryItemAction{
			interfaceID:       interfaceID,
			item:              item,
			secondaryActionID: secondaryActionID,
		},
	}
}

// DeferEquipItem sets the player's pending action to equip an inventory item. This will overwrite any previously
// deferred action.
func (pe *playerEntity) DeferEquipItem(item *model.Item, interfaceID int) {
	pe.deferredAction = &pendingAction{
		actionType: pendingActionEquipItem,
		equipItemAction: &equipItemAction{
			interfaceID: interfaceID,
			item:        item,
		},
	}
}

// DeferUnequipItem sets the player's pending action to equip an inventory item. This will overwrite any previously
// deferred action.
func (pe *playerEntity) DeferUnequipItem(item *model.Item, interfaceID int, slotType model.EquipmentSlotType) {
	pe.deferredAction = &pendingAction{
		actionType: pendingActionUnequipItem,
		unequipItemAction: &unequipItemAction{
			interfaceID: interfaceID,
			item:        item,
			slotType:    slotType,
		},
	}
}

// MoveDirection returns the direction the player is currently moving in. If the player is not moving, then
// model.DirectionNone will be returned.
func (pe *playerEntity) MoveDirection() model.Direction {
	if !pe.Walking() {
		return model.DirectionNone
	}

	return model.DirectionFromDelta(pe.path[pe.nextPathIdx].Sub(pe.player.GlobalPos.To2D()))
}

// Walking determines if the player is walking to a destination.
func (pe *playerEntity) Walking() bool {
	return pe.nextPathIdx < len(pe.path)
}

// PlanEvent adds a scheduled event to this player's queue and resets the event timer.
func (pe *playerEntity) PlanEvent(e *Event) {
	pe.scheduler.Plan(e)
}
