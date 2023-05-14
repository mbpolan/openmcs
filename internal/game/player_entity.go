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
	lastChatMessage     *model.ChatMessage
	lastChatTime        time.Time
	chatHighWater       time.Time
	tabInterfaces       map[model.ClientTab]int
	teleportGlobal      *model.Vector3D
	privateMessageID    int
	regionOrigin        model.Vector2D
	appearanceChanged   bool
	nextStatusBroadcast *playerStatusBroadcast
	nextUpdate          *response.PlayerUpdateResponse
	deferredActions     []*Action
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

// TickDeferredActions decrements the tick delay on all deferred actions and returns a slice of actions that are ready
// for processing.
func (pe *playerEntity) TickDeferredActions() []*Action {
	var expired []*Action

	for _, deferred := range pe.deferredActions {
		if deferred.TickDelay >= 1 {
			deferred.TickDelay--
		} else {
			expired = append(expired, deferred)
		}
	}

	return expired
}

// RemoveDeferredAction removes a deferred action.
func (pe *playerEntity) RemoveDeferredAction(action *Action) {
	for i, deferred := range pe.deferredActions {
		if deferred == action {
			pe.deferredActions = append(pe.deferredActions[:i], pe.deferredActions[i+1:]...)
			return
		}
	}
}

// DeferTakeGroundItemAction sets the player's pending action to pick up a specific ground Item at a position, in
// global coordinates. This will overwrite any previously deferred action.
func (pe *playerEntity) DeferTakeGroundItemAction(item *model.Item, globalPos model.Vector3D) {
	pe.deferredActions = append(pe.deferredActions, &Action{
		ActionType: pendingActionTakeGroundItem,
		TickDelay:  1,
		TakeGroundItem: &TakeGroundItemAction{
			GlobalPos: globalPos,
			Item:      item,
		},
	})
}

// DeferDropInventoryItem sets the player's pending action to drop an inventory Item. This will overwrite any previously
// deferred action.
func (pe *playerEntity) DeferDropInventoryItem(item *model.Item, interfaceID, secondaryActionID int) {
	pe.deferredActions = append(pe.deferredActions, &Action{
		ActionType: pendingActionDropInventoryItem,
		TickDelay:  1,
		DropInventoryItemAction: &DropInventoryItemAction{
			InterfaceID:       interfaceID,
			Item:              item,
			SecondaryActionID: secondaryActionID,
		},
	})
}

// DeferEquipItem sets the player's pending action to equip an inventory Item. This will overwrite any previously
// deferred action.
func (pe *playerEntity) DeferEquipItem(item *model.Item, interfaceID int) {
	pe.deferredActions = append(pe.deferredActions, &Action{
		ActionType: pendingActionEquipItem,
		TickDelay:  1,
		EquipItemAction: &EquipItemAction{
			InterfaceID: interfaceID,
			Item:        item,
		},
	})
}

// DeferUnequipItem sets the player's pending action to equip an inventory Item. This will overwrite any previously
// deferred action.
func (pe *playerEntity) DeferUnequipItem(item *model.Item, interfaceID int, slotType model.EquipmentSlotType) {
	pe.deferredActions = append(pe.deferredActions, &Action{
		ActionType: pendingActionUnequipItem,
		TickDelay:  1,
		UnequipItemAction: &UnequipItemAction{
			InterfaceID: interfaceID,
			Item:        item,
			SlotType:    slotType,
		},
	})
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
