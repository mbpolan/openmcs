package game

import (
	"github.com/google/uuid"
	"github.com/mbpolan/openmcs/internal/model"
	"github.com/mbpolan/openmcs/internal/network/response"
	"time"
)

type EventType int

const (
	// EventSendResponse sends a generic response to the client.
	EventSendResponse EventType = iota
	// EventSendManyResponses sends multiple, generic responses to the client.
	EventSendManyResponses
	// EventCheckIdle is a recurring, scheduled check for player inactivity.
	EventCheckIdle
	// EventCheckIdleImmediate is a one-off check for player inactivity.
	EventCheckIdleImmediate
	// EventUpdateTabInterfaces updates all the player's client tab interfaces.
	EventUpdateTabInterfaces
	// EventSendInventory sends a complete player inventory update.
	EventSendInventory
	// EventSwapInventoryItem rearranges an item in a player's inventory.
	EventSwapInventoryItem
	// EventFriendList sends a player's entire friends list.
	EventFriendList
	// EventSkills sends data about all the player's skills.
	EventSkills
	// EventRemoveExpiredGroundItem removes a ground item on a tile after it has expired.
	EventRemoveExpiredGroundItem
)

// Event is an action that the game server should take at a specified time.
type Event struct {
	Type         EventType
	Schedule     time.Time
	Responses    []response.Response
	InstanceUUID uuid.UUID
	GlobalPos    model.Vector3D
	FromSlot     int
	ToSlot       int
}

// NewEventWithType creates an event with a specific type that should be processed at the provided time.
func NewEventWithType(eventType EventType, when time.Time) *Event {
	return &Event{
		Type:     eventType,
		Schedule: when,
	}
}

// NewSendResponseEvent creates an event that will send a game state update to a player at the provided time.
func NewSendResponseEvent(resp response.Response, when time.Time) *Event {
	return &Event{
		Type:      EventSendResponse,
		Schedule:  when,
		Responses: []response.Response{resp},
	}
}

// NewSendMultipleResponsesEvent creates an event that will send multiple responses to a player.
func NewSendMultipleResponsesEvent(responses []response.Response, when time.Time) *Event {
	return &Event{
		Type:      EventSendManyResponses,
		Schedule:  when,
		Responses: responses,
	}
}
