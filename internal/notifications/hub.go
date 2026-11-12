package notifications

import (
	"sync"
	"time"
)

// Event is one notification delivered to web clients (SSE) and mobile clients
// (FCM). The JSON shape is the wire contract: type/title/body/url/time plus an
// optional data bag of string values.
type Event struct {
	Type  string            `json:"type"`
	Title string            `json:"title"`
	Body  string            `json:"body"`
	URL   string            `json:"url"`
	Data  map[string]string `json:"data,omitempty"`
	Time  time.Time         `json:"time"`
}

// subscriberBuffer is how many events a slow web client may fall behind before
// its events are dropped. Publish must never block a request handler.
const subscriberBuffer = 16

// Hub fans events out to per-user SSE subscribers. A user may have several
// open browser tabs, so subscriptions are stored per connection.
type Hub struct {
	mu     sync.RWMutex
	subs   map[uint]map[int]chan Event
	nextID int
}

func NewHub() *Hub {
	return &Hub{subs: make(map[uint]map[int]chan Event)}
}

// Register subscribes a connection for userID. The returned cancel function is
// idempotent and must be called when the connection ends.
func (h *Hub) Register(userID uint) (<-chan Event, func()) {
	ch := make(chan Event, subscriberBuffer)

	h.mu.Lock()
	if h.subs[userID] == nil {
		h.subs[userID] = make(map[int]chan Event)
	}
	h.nextID++
	id := h.nextID
	h.subs[userID][id] = ch
	h.mu.Unlock()

	var once sync.Once
	cancel := func() {
		once.Do(func() {
			h.mu.Lock()
			if conns, ok := h.subs[userID]; ok {
				delete(conns, id)
				if len(conns) == 0 {
					delete(h.subs, userID)
				}
			}
			h.mu.Unlock()
		})
	}
	return ch, cancel
}

// Publish delivers ev to every subscriber of the given users. Delivery is
// best-effort: a subscriber whose buffer is full misses the event instead of
// stalling the publisher.
func (h *Hub) Publish(userIDs []uint, ev Event) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for _, userID := range userIDs {
		for _, ch := range h.subs[userID] {
			select {
			case ch <- ev:
			default:
			}
		}
	}
}
