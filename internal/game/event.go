package game

import (
	"github.com/mbpolan/openmcs/internal/network/responses"
	"time"
)

type EventType int

const (
	EventPlayerUpdate EventType = iota
	EventSendResponse
)

// Event is an action that the game server should take at a specified time.
type Event struct {
	Type     EventType
	Schedule time.Time
	Response responses.Response
}

// NewEventWithType creates an event with a specific type that should be processed at the provided time.
func NewEventWithType(eventType EventType, when time.Time) *Event {
	return &Event{
		Type:     eventType,
		Schedule: when,
	}
}

// NewSendResponseEvent creates an event that will send a game state update to a player at the provided time.
func NewSendResponseEvent(response responses.Response, when time.Time) *Event {
	return &Event{
		Type:     EventSendResponse,
		Schedule: when,
		Response: response,
	}
}
