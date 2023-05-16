package game

import (
	"github.com/mbpolan/openmcs/internal/model"
	"github.com/mbpolan/openmcs/internal/network"
	"github.com/mbpolan/openmcs/internal/network/response"
	"sync"
	"time"
)

// PlayerEntity represents a player and their state while they are logged into the game world.
type PlayerEntity struct {
	lastInteraction     time.Time
	player              *model.Player
	tracking            map[int]*PlayerEntity
	changeChan          chan bool
	doneChan            chan bool
	outChan             chan response.Response
	path                []model.Vector2D
	nextPathIdx         int
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
func newPlayerEntity(p *model.Player, w *network.ProtocolWriter) *PlayerEntity {
	changeChan := make(chan bool)

	return &PlayerEntity{
		lastInteraction:  time.Now(),
		player:           p,
		tracking:         map[int]*PlayerEntity{},
		changeChan:       changeChan,
		doneChan:         make(chan bool, 1),
		outChan:          make(chan response.Response, 50),
		privateMessageID: 1,
		writer:           w,
	}
}

// MarkStatusBroadcast marks that this player's online/offline status should be broadcast to everyone.
func (pe *PlayerEntity) MarkStatusBroadcast() {
	if pe.nextStatusBroadcast == nil {
		pe.nextStatusBroadcast = &playerStatusBroadcast{}
	}

	pe.nextStatusBroadcast.targets = nil
}

// MarkStatusBroadcastTarget adds a target to receive this player's online/offline status.
func (pe *PlayerEntity) MarkStatusBroadcastTarget(target string) {
	if pe.nextStatusBroadcast == nil {
		pe.nextStatusBroadcast = &playerStatusBroadcast{}
	}

	pe.nextStatusBroadcast.targets = append(pe.nextStatusBroadcast.targets, target)
}

// MoveDirection returns the direction the player is currently moving in. If the player is not moving, then
// model.DirectionNone will be returned.
func (pe *PlayerEntity) MoveDirection() model.Direction {
	if !pe.Walking() {
		return model.DirectionNone
	}

	return model.DirectionFromDelta(pe.path[pe.nextPathIdx].Sub(pe.player.GlobalPos.To2D()))
}

// Walking determines if the player is walking to a destination.
func (pe *PlayerEntity) Walking() bool {
	return pe.nextPathIdx < len(pe.path)
}

// Send adds one or more responses that will be sent to the player.
func (pe *PlayerEntity) Send(responses ...response.Response) {
	for _, resp := range responses {
		select {
		case pe.outChan <- resp:

		default:
			// write to the done chan since this player is too far behind on responses
			pe.Drop()
			return
		}
	}
}

// Drop flags that this player should be disconnected and no more responses should be sent to the client.
func (pe *PlayerEntity) Drop() {
	select {
	case pe.doneChan <- true:
	default:
	}
}

// TickDeferredActions decrements the tick delay on all deferred actions and returns a slice of actions that are ready
// for processing.
func (pe *PlayerEntity) TickDeferredActions() []*Action {
	var expired []*Action

	for _, deferred := range pe.deferredActions {
		if deferred.TickDelay >= 1 {
			deferred.TickDelay--
		}

		if deferred.TickDelay == 0 {
			expired = append(expired, deferred)
		}
	}

	return expired
}

// RemoveDeferredAction removes a deferred action.
func (pe *PlayerEntity) RemoveDeferredAction(action *Action) {
	for i, deferred := range pe.deferredActions {
		if deferred == action {
			pe.deferredActions = append(pe.deferredActions[:i], pe.deferredActions[i+1:]...)
			return
		}
	}
}

// DeferMoveInventoryItem plans an action to move an item in the player's inventory from one slot to another.
func (pe *PlayerEntity) DeferMoveInventoryItem(fromSlot, toSlot int) {
	pe.deferredActions = append(pe.deferredActions, &Action{
		ActionType: ActionMoveInventoryItem,
		TickDelay:  1,
		MoveInventoryItemAction: &MoveInventoryItemAction{
			FromSlot: fromSlot,
			ToSlot:   toSlot,
		},
	})
}

// DeferSendServerMessage plans an action to send a player a server message.
func (pe *PlayerEntity) DeferSendServerMessage(message string) {
	pe.deferredActions = append(pe.deferredActions, &Action{
		ActionType: ActionSendServerMessage,
		TickDelay:  0,
		ServerMessageAction: &ServerMessageAction{
			Message: message,
		},
	})
}

// DeferSendSkills plans an action to send a player their current skill stats.
func (pe *PlayerEntity) DeferSendSkills() {
	pe.deferredActions = append(pe.deferredActions, &Action{
		ActionType: ActionSendSkills,
		TickDelay:  1,
	})
}

// DeferSendInterfaces plans an action to send a player the client tab interface to display.
func (pe *PlayerEntity) DeferSendInterfaces() {
	pe.deferredActions = append(pe.deferredActions, &Action{
		ActionType: ActionSendInterfaces,
		TickDelay:  1,
	})
}

// DeferSendModes plans an action to send a player their current chat modes.
func (pe *PlayerEntity) DeferSendModes() {
	pe.deferredActions = append(pe.deferredActions, &Action{
		ActionType: ActionSendModes,
		TickDelay:  1,
	})
}

// DeferSendEquipment plans an action to send a player their current equipped items.
func (pe *PlayerEntity) DeferSendEquipment() {
	pe.deferredActions = append(pe.deferredActions, &Action{
		ActionType: ActionSendEquipment,
		TickDelay:  1,
	})
}

// DeferSendInventory plans an action to send a player their current inventory items.
func (pe *PlayerEntity) DeferSendInventory() {
	pe.deferredActions = append(pe.deferredActions, &Action{
		ActionType: ActionSendInventory,
		TickDelay:  1,
	})
}

// DeferSendFriendList plans an action to send a player their friend list and each friend's status.
func (pe *PlayerEntity) DeferSendFriendList() {
	pe.deferredActions = append(pe.deferredActions, &Action{
		ActionType: ActionSendFriendList,
		TickDelay:  1,
	})
}

// DeferSendIgnoreList plans an action to send a player their ignore list.
func (pe *PlayerEntity) DeferSendIgnoreList() {
	pe.deferredActions = append(pe.deferredActions, &Action{
		ActionType: ActionSendIgnoreList,
		TickDelay:  1,
	})
}

// DeferTakeGroundItemAction sets the player's pending action to pick up a specific ground Item at a position, in
// global coordinates. This will overwrite any previously deferred action.
func (pe *PlayerEntity) DeferTakeGroundItemAction(item *model.Item, globalPos model.Vector3D) {
	pe.deferredActions = append(pe.deferredActions, &Action{
		ActionType: ActionTakeGroundItem,
		TickDelay:  1,
		TakeGroundItem: &TakeGroundItemAction{
			GlobalPos: globalPos,
			Item:      item,
		},
	})
}

// DeferDropInventoryItem sets the player's pending action to drop an inventory Item. This will overwrite any previously
// deferred action.
func (pe *PlayerEntity) DeferDropInventoryItem(item *model.Item, interfaceID, secondaryActionID int) {
	pe.deferredActions = append(pe.deferredActions, &Action{
		ActionType: ActionDropInventoryItem,
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
func (pe *PlayerEntity) DeferEquipItem(item *model.Item, interfaceID int) {
	pe.deferredActions = append(pe.deferredActions, &Action{
		ActionType: ActionEquipItem,
		TickDelay:  1,
		EquipItemAction: &EquipItemAction{
			InterfaceID: interfaceID,
			Item:        item,
		},
	})
}

// DeferUnequipItem sets the player's pending action to equip an inventory Item. This will overwrite any previously
// deferred action.
func (pe *PlayerEntity) DeferUnequipItem(item *model.Item, interfaceID int, slotType model.EquipmentSlotType) {
	pe.deferredActions = append(pe.deferredActions, &Action{
		ActionType: ActionUnequipItem,
		TickDelay:  1,
		UnequipItemAction: &UnequipItemAction{
			InterfaceID: interfaceID,
			Item:        item,
			SlotType:    slotType,
		},
	})
}
