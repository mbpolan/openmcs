package game

import (
	"sync"
	"time"
)

// Scheduler provides a thread-safe queue of events that can be planned on a time basis.
type Scheduler struct {
	events []*Event
	mu     sync.Mutex
}

// NewScheduler creates a new event scheduler.
func NewScheduler() *Scheduler {
	return &Scheduler{}
}

// TimeUntil returns the time left until the next event should be processed. If there are no events, then a time
// duration in the far future will be returned.
func (s *Scheduler) TimeUntil() time.Duration {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.events) == 0 {
		return time.Hour
	}

	delta := time.Until(s.events[0].Schedule)
	return delta
}

// Next returns the next event that should be processed. If there are no events, nil will be returned instead.
func (s *Scheduler) Next() *Event {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.events) == 0 {
		return nil
	}

	e := s.events[0]
	s.events = s.events[1:]
	return e
}

// Plan schedules an event for later processing.
func (s *Scheduler) Plan(e *Event) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// find the position where this event should be
	pos := 0
	for i, other := range s.events {
		if other.Schedule.After(e.Schedule) {
			pos = i
			break
		}

		pos++
	}

	// no events in the queue or at the end of the queue
	if len(s.events) == pos {
		s.events = append(s.events, e)
		return
	}

	// insert at a specific position
	s.events = append(s.events[:pos+1], s.events[pos:]...)
	s.events[pos] = e
}
