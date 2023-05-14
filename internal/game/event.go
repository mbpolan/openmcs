package game

import (
	"github.com/google/uuid"
	"github.com/mbpolan/openmcs/internal/model"
	"time"
)

type EventType int

const (
	// EventRemoveExpiredGroundItem removes a ground Item on a tile after it has expired.
	EventRemoveExpiredGroundItem EventType = iota
)

// Event is an action that the game server should take at a specified time.
type Event struct {
	Type         EventType
	Schedule     time.Time
	InstanceUUID uuid.UUID
	GlobalPos    model.Vector3D
}

// NewEventWithType creates an event with a specific type that should be processed at the provided time.
func NewEventWithType(eventType EventType, when time.Time) *Event {
	return &Event{
		Type:     eventType,
		Schedule: when,
	}
}
