package game

import (
	"github.com/mbpolan/openmcs/internal/model"
	"github.com/mbpolan/openmcs/internal/network"
	"github.com/mbpolan/openmcs/internal/network/response"
	"sync"
	"time"
)

// playerEntity represents a player and their state while they are logged into the game world.
type playerEntity struct {
	lastInteraction time.Time
	player          *model.Player
	tracking        map[int]*playerEntity
	resetChan       chan bool
	doneChan        chan bool
	path            []model.Vector2D
	nextPathIdx     int
	scheduler       *Scheduler
	writer          *network.ProtocolWriter
	lastWalkTime    time.Time
	lastChatMessage *model.ChatMessage
	lastChatTime    time.Time
	chatHighWater   time.Time
	tabInterfaces   map[model.ClientTab]int
	nextUpdate      *response.PlayerUpdateResponse
	mu              sync.Mutex
}

// newPlayerEntity creates a new player entity.
func newPlayerEntity(p *model.Player, w *network.ProtocolWriter) *playerEntity {
	return &playerEntity{
		lastInteraction: time.Now(),
		player:          p,
		tracking:        map[int]*playerEntity{},
		resetChan:       make(chan bool),
		doneChan:        make(chan bool, 1),
		scheduler:       NewScheduler(),
		writer:          w,
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
	pe.resetChan <- true
}
