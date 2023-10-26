package game

import (
	"github.com/mbpolan/openmcs/internal/logger"
	"github.com/mbpolan/openmcs/internal/model"
	"github.com/mbpolan/openmcs/internal/network"
	"github.com/mbpolan/openmcs/internal/network/response"
	"sync"
	"time"
)

// maxQueueSize is the maximum amount of responses queued for a player to receive.
const maxQueueSize = 100

// playerEntity represents a player and their state while they are logged into the game world.
type playerEntity struct {
	index               int
	lastInteraction     time.Time
	player              *model.Player
	tracking            map[int]*playerEntity
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
	privateMessageID    int
	regionOrigin        model.Vector2D
	appearanceChanged   bool
	lastAnimations      map[model.AnimationID]int
	nextStatusBroadcast *playerStatusBroadcast
	nextUpdate          *response.PlayerUpdateResponse
	statRegenTicks      map[model.SkillType]int
	deferredActions     []*Action
	mu                  sync.Mutex
	animationTicks      int
	graphicTicks        int
	graphicApplied      bool
	isLowMemory         bool
}

type playerStatusBroadcast struct {
	targets []string
}

// newPlayerEntity creates a new player entity.
func newPlayerEntity(p *model.Player, w *network.ProtocolWriter) *playerEntity {
	changeChan := make(chan bool)

	return &playerEntity{
		animationTicks:   -1,
		lastInteraction:  time.Now(),
		player:           p,
		tracking:         map[int]*playerEntity{},
		changeChan:       changeChan,
		doneChan:         make(chan bool, 1),
		outChan:          make(chan response.Response, maxQueueSize),
		privateMessageID: 1,
		statRegenTicks:   map[model.SkillType]int{},
		tabInterfaces:    map[model.ClientTab]int{},
		writer:           w,
	}
}

// Animating returns true if the player has an ongoing animation, false if not.
func (pe *playerEntity) Animating() bool {
	if pe.lastAnimations == nil {
		return false
	}

	return true
}

// AnimationID returns the ID of the animation the player is currently performing.
func (pe *playerEntity) AnimationID() int {
	return pe.player.Appearance.Animations[model.AnimationStand]
}

// SetAnimation sets an animation the player's client should start performing, taking precedence over the default
// animation. This will also flag the player's appearance as changed. The tickDuration specifies after how many game
// ticks the animation will be stopped.
func (pe *playerEntity) SetAnimation(animationID, tickDuration int) {
	// save the player's current animations
	pe.lastAnimations = map[model.AnimationID]int{}
	for k, v := range pe.player.Appearance.Animations {
		pe.lastAnimations[k] = v
	}

	// set the new animation
	pe.player.Appearance.Animations[model.AnimationStand] = animationID
	pe.animationTicks = tickDuration
	pe.appearanceChanged = true
}

// ClearAnimation removes the current animation for the player. This will also flag the player's appearance as changed.
func (pe *playerEntity) ClearAnimation() {
	// restore the player's previous animations
	for k, v := range pe.lastAnimations {
		pe.player.Appearance.Animations[k] = v
	}

	pe.lastAnimations = nil
	pe.animationTicks = -1
	pe.appearanceChanged = true
}

// HasGraphic returns true if a graphic should be displayed along with the player model.
func (pe *playerEntity) HasGraphic() bool {
	return pe.player.Appearance.GraphicID != -1
}

// GraphicID returns the ID of the graphic the player model is currently displaying.
func (pe *playerEntity) GraphicID() int {
	return pe.player.Appearance.GraphicID
}

// GraphicHeight returns the height offset from the ground where the player model graphic should be rendered. If no
// graphic is set on the player model, the return value from this method is undefined.
func (pe *playerEntity) GraphicHeight() int {
	return pe.player.Appearance.GraphicHeight
}

// GraphicDelay returns the number of client-side ticks to wait before the player model graphic should be initially
// rendered. If no graphic is set on the player model, the return value from this method is undefined.
func (pe *playerEntity) GraphicDelay() int {
	return pe.player.Appearance.GraphicDelay
}

// SetGraphic sets the graphic to display along with the player's model. This will also flag the player's appearance as
// changed. The delay is how many client-side ticks (frames) to wait before displaying the graphic. The tickDuration
// specifies after how many game ticks the graphic will be cleared.
func (pe *playerEntity) SetGraphic(graphicID, height, delay, tickDuration int) {
	pe.player.Appearance.GraphicID = graphicID
	pe.player.Appearance.GraphicHeight = height
	pe.player.Appearance.GraphicDelay = delay
	pe.graphicApplied = false
	pe.graphicTicks = tickDuration
	pe.appearanceChanged = true
}

// ClearGraphic removes the current graphic for the player. This will also flag the player's appearance as changed.
func (pe *playerEntity) ClearGraphic() {
	pe.player.Appearance.GraphicID = -1
	pe.player.Appearance.GraphicHeight = 0
	pe.player.Appearance.GraphicDelay = 0
	pe.graphicTicks = -1
	pe.appearanceChanged = true
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

// Moving determines if the player is walking or running to a destination.
func (pe *playerEntity) Moving() bool {
	return pe.nextPathIdx < len(pe.path)
}

// Send adds one or more responses that will be sent to the player.
func (pe *playerEntity) Send(responses ...response.Response) {
	for _, resp := range responses {
		select {
		case pe.outChan <- resp:

		default:
			// write to the done chan since this player is too far behind on responses
			logger.Warnf("overflow in player %d response queue", pe.player.ID)
			pe.Drop()
			return
		}
	}
}

// Drop flags that this player should be disconnected and no more responses should be sent to the client.
func (pe *playerEntity) Drop() {
	select {
	case pe.doneChan <- true:
	default:
	}
}

// TickDeferredActions decrements the tick delay on all deferred actions and returns a slice of actions that are ready
// for processing.
func (pe *playerEntity) TickDeferredActions() []*Action {
	var expired []*Action

	for _, deferred := range pe.deferredActions {
		if deferred.TickDelay >= 1 {
			deferred.TickDelay--
		}

		// stop iterating once we reach an action that is still pending
		if deferred.TickDelay == 0 {
			expired = append(expired, deferred)
		} else {
			break
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

// DeferMoveInventoryItem plans an action to move an item in the player's inventory from one slot to another.
func (pe *playerEntity) DeferMoveInventoryItem(fromSlot, toSlot int) {
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
func (pe *playerEntity) DeferSendServerMessage(message string) {
	pe.deferredActions = append(pe.deferredActions, &Action{
		ActionType: ActionSendServerMessage,
		TickDelay:  0,
		ServerMessageAction: &ServerMessageAction{
			Message: message,
		},
	})
}

// DeferSendSkills plans an action to send a player their current skill stats.
func (pe *playerEntity) DeferSendSkills(skillTypes []model.SkillType) {
	pe.deferredActions = append(pe.deferredActions, &Action{
		ActionType: ActionSendSkills,
		TickDelay:  1,
		SendSkillsAction: &SendSkillsAction{
			SkillTypes: skillTypes,
		},
	})
}

// DeferSendModes plans an action to send a player their current chat modes.
func (pe *playerEntity) DeferSendModes() {
	pe.deferredActions = append(pe.deferredActions, &Action{
		ActionType: ActionSendModes,
		TickDelay:  1,
	})
}

// DeferSendEquipment plans an action to send a player their current equipped items.
func (pe *playerEntity) DeferSendEquipment() {
	pe.deferredActions = append(pe.deferredActions, &Action{
		ActionType: ActionSendEquipment,
		TickDelay:  1,
	})
}

// DeferSendInventory plans an action to send a player their current inventory items.
func (pe *playerEntity) DeferSendInventory() {
	pe.deferredActions = append(pe.deferredActions, &Action{
		ActionType: ActionSendInventory,
		TickDelay:  1,
	})
}

// DeferSendFriendList plans an action to send a player their friend list and each friend's status.
func (pe *playerEntity) DeferSendFriendList() {
	pe.deferredActions = append(pe.deferredActions, &Action{
		ActionType: ActionSendFriendList,
		TickDelay:  1,
	})
}

// DeferSendIgnoreList plans an action to send a player their ignore list.
func (pe *playerEntity) DeferSendIgnoreList() {
	pe.deferredActions = append(pe.deferredActions, &Action{
		ActionType: ActionSendIgnoreList,
		TickDelay:  1,
	})
}

// DeferTakeGroundItemAction plans an action to pick up a specific ground item at a position, in global coordinates.
func (pe *playerEntity) DeferTakeGroundItemAction(item *model.Item, globalPos model.Vector3D) {
	pe.deferredActions = append(pe.deferredActions, &Action{
		ActionType: ActionTakeGroundItem,
		TickDelay:  1,
		TakeGroundItem: &TakeGroundItemAction{
			GlobalPos: globalPos,
			Item:      item,
		},
	})
}

// DeferDropInventoryItem plans an action to drop an inventory item.
func (pe *playerEntity) DeferDropInventoryItem(item *model.Item, interfaceID, secondaryActionID int) {
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

// DeferEquipItem plans an action to equip an inventory item.
func (pe *playerEntity) DeferEquipItem(item *model.Item, interfaceID int) {
	pe.deferredActions = append(pe.deferredActions, &Action{
		ActionType: ActionEquipItem,
		TickDelay:  1,
		EquipItemAction: &EquipItemAction{
			InterfaceID: interfaceID,
			Item:        item,
		},
	})
}

// DeferUnequipItem plans an action to unequip an inventory item.
func (pe *playerEntity) DeferUnequipItem(item *model.Item, interfaceID int, slotType model.EquipmentSlotType) {
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

// DeferShowInterface plans an action to show an interface.
func (pe *playerEntity) DeferShowInterface(interfaceID int) {
	pe.deferredActions = append(pe.deferredActions, &Action{
		ActionType: ActionShowInterface,
		TickDelay:  1,
		ShowInterfaceAction: &ShowInterfaceAction{
			InterfaceID: interfaceID,
		},
	})
}

// DeferHideInterfaces plans an action to hide all interfaces.
func (pe *playerEntity) DeferHideInterfaces() {
	pe.deferredActions = append(pe.deferredActions, &Action{
		ActionType: ActionHideInterfaces,
		TickDelay:  1,
	})
}

// DeferDoInterfaceAction plans an action to handle an interaction with an interface.
func (pe *playerEntity) DeferDoInterfaceAction(parent, actor *model.Interface) {
	pe.deferredActions = append(pe.deferredActions, &Action{
		ActionType: ActionDoInterfaceAction,
		DoInterfaceAction: &DoInterfaceAction{
			Parent: parent,
			Actor:  actor,
		},
	})
}

// DeferTeleportPlayer plans an action to teleport the player to a position, in global coordinates.
func (pe *playerEntity) DeferTeleportPlayer(globalPos model.Vector3D) {
	pe.deferredActions = append(pe.deferredActions, &Action{
		ActionType: ActionTeleportPlayer,
		TickDelay:  1,
		TeleportPlayerAction: &TeleportPlayerAction{
			GlobalPos: globalPos,
		},
	})
}

// DeferCastSpellOnItem plans an action to cast a spell on a player's inventory item.
func (pe *playerEntity) DeferCastSpellOnItem(slotID, itemID, inventoryInterfaceID, spellInterfaceID int) {
	pe.deferredActions = append(pe.deferredActions, &Action{
		ActionType: ActionCastSpellOnItem,
		TickDelay:  1,
		CastSpellOnItemAction: &CastSpellOnItemAction{
			SlotID:               slotID,
			ItemID:               itemID,
			InventoryInterfaceID: inventoryInterfaceID,
			SpellInterfaceID:     spellInterfaceID,
		},
	})
}

// DeferExperienceGrant plans an action to grant the player experience in a skill.
func (pe *playerEntity) DeferExperienceGrant(skillType model.SkillType, experience float64) {
	pe.planAction(&Action{
		ActionType: ActionExperienceGrant,
		TickDelay:  1,
		ExperienceGrantAction: &ExperienceGrantAction{
			SkillType:  skillType,
			Experience: experience,
		},
	}, ActionPriorityHigh)
}

// DeferSendChangeEvent plans an action to communicate that an attribute on the player has changed.
func (pe *playerEntity) DeferSendChangeEvent(event model.PlayerChangeEvent) {
	pe.planAction(&Action{
		ActionType: ActionSendChangeEvent,
		TickDelay:  1,
		PlayerChangeEventAction: &PlayerChangeEventAction{
			Event: event,
		},
	})
}

// DeferSendRunEnergy plans an action send a player their current run energy.
func (pe *playerEntity) DeferSendRunEnergy() {
	pe.planAction(&Action{
		ActionType: ActionSendRunEnergy,
		TickDelay:  1,
	})
}

// DeferSendWeight plans an action send a player their current weight.
func (pe *playerEntity) DeferSendWeight() {
	pe.planAction(&Action{
		ActionType: ActionSendWeight,
		TickDelay:  1,
	})
}

// DeferPlayMusic plans an action send a player's client a music track to play.
func (pe *playerEntity) DeferPlayMusic(musicID int) {
	pe.planAction(&Action{
		ActionType: ActionPlayMusic,
		TickDelay:  1,
		PlayMusicAction: &PlayMusicAction{
			MusicID: musicID,
		},
	})
}

// DeferActionCompletion plans an artificial delay to indicate the player is occupied with an ongoing action.
func (pe *playerEntity) DeferActionCompletion(tickDuration int) {
	pe.planAction(&Action{
		ActionType: ActionDelayCurrent,
		TickDelay:  uint(tickDuration),
	}, ActionPriorityHigh)
}

// planAction plans a deferred action with a priority.
func (pe *playerEntity) planAction(action *Action, priority ...ActionPriority) {
	if len(priority) == 0 || priority[0] == ActionPriorityNormal {
		pe.deferredActions = append(pe.deferredActions, action)
		return
	}

	pe.deferredActions = append([]*Action{action}, pe.deferredActions...)
}
