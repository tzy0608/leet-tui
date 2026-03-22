package pubsub

import "sync"

// Event represents a named event with optional data.
type Event struct {
	Type string
	Data any
}

// Common event types
const (
	EventProblemsLoaded    = "problems.loaded"
	EventProblemSelected   = "problem.selected"
	EventSubmissionResult  = "submission.result"
	EventReviewCompleted   = "review.completed"
	EventPlanActivated     = "plan.activated"
	EventSyncStarted      = "sync.started"
	EventSyncCompleted    = "sync.completed"
	EventSyncError        = "sync.error"
)

type Handler func(Event)

type Broker struct {
	mu       sync.RWMutex
	subs     map[string][]Handler
	nextID   int
}

func New() *Broker {
	return &Broker{
		subs: make(map[string][]Handler),
	}
}

func (b *Broker) Subscribe(eventType string, handler Handler) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.subs[eventType] = append(b.subs[eventType], handler)
}

func (b *Broker) Publish(event Event) {
	b.mu.RLock()
	handlers := make([]Handler, len(b.subs[event.Type]))
	copy(handlers, b.subs[event.Type])
	b.mu.RUnlock()

	for _, h := range handlers {
		h(event)
	}
}
