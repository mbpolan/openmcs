package game

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func Test_Scheduler_TimeUntilNoEvents(t *testing.T) {
	s := NewScheduler()

	next := s.TimeUntil()

	assert.True(t, next >= time.Hour)
}

func Test_Scheduler_OneEvent(t *testing.T) {
	s := NewScheduler()

	s.Plan(NewEventWithType(EventPlayerUpdate, time.Now()))

	assert.Equal(t, 1, len(s.events))
}

func Test_Scheduler_EventSequence(t *testing.T) {
	s := NewScheduler()

	s.Plan(NewEventWithType(EventPlayerUpdate, time.Now()))
	s.Plan(NewEventWithType(EventSendResponse, time.Now().Add(1*time.Minute)))

	assert.Equal(t, 2, len(s.events))
	assert.Equal(t, EventPlayerUpdate, s.events[0].Type)
	assert.Equal(t, EventSendResponse, s.events[1].Type)
}

func Test_Scheduler_TimeUntilSoonestEvent(t *testing.T) {
	s := NewScheduler()

	first := time.Now().Add(-1 * time.Second)
	second := time.Now().Add(1 * time.Minute)

	s.Plan(NewEventWithType(EventPlayerUpdate, first))
	s.Plan(NewEventWithType(EventSendResponse, second))

	// first event is ready to process
	next := s.TimeUntil()
	assert.True(t, next < 0)

	_ = s.Next()

	// second event still has some time remaining
	next = s.TimeUntil()
	assert.True(t, next > 0)
}
