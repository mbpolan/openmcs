package game

import (
	"github.com/mbpolan/openmcs/internal/network/response"
	"time"
)

type EventType int

const (
	// EventPlayerUpdate indicates a game state update should be sent to the player.
	EventPlayerUpdate EventType = iota
	// EventSendResponse sends a generic response to the client.
	EventSendResponse
	// EventCheckIdle is a recurring, scheduled check for player inactivity.
	EventCheckIdle
	// EventCheckIdleImmediate is a one-off check for player inactivity.
	EventCheckIdleImmediate
	// EventUpdateTabInterfaces updates all of the player's client tab interfaces.
	EventUpdateTabInterfaces
	// EventFriendStatusUpdate conveys a change to a single friends list player status.
	EventFriendStatusUpdate
	// EventFriendList sends a player's entire friends list.
	EventFriendList
)

// Event is an action that the game server should take at a specified time.
type Event struct {
	Type     EventType
	Schedule time.Time
	Response response.Response
}

// NewEventWithType creates an event with a specific type that should be processed at the provided time.
func NewEventWithType(eventType EventType, when time.Time) *Event {
	return &Event{
		Type:     eventType,
		Schedule: when,
	}
}

// NewSendResponseEvent creates an event that will send a game state update to a player at the provided time.
func NewSendResponseEvent(response response.Response, when time.Time) *Event {
	return &Event{
		Type:     EventSendResponse,
		Schedule: when,
		Response: response,
	}
}
