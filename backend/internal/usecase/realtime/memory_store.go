package realtime

import (
	"context"
	"sync"
	"time"
)

type memoryTicket struct {
	ticket    Ticket
	expiresAt time.Time
}

// MemoryTicketStore は Redis がない環境でプロセス内にチケットを保存します
type MemoryTicketStore struct {
	mu      sync.Mutex
	tickets map[string]memoryTicket
}

func NewMemoryTicketStore() *MemoryTicketStore {
	return &MemoryTicketStore{tickets: map[string]memoryTicket{}}
}

func (s *MemoryTicketStore) Save(_ context.Context, hash string, ticket Ticket, ttl time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now()
	for key, t := range s.tickets {
		if now.After(t.expiresAt) {
			delete(s.tickets, key)
		}
	}
	s.tickets[hash] = memoryTicket{ticket: ticket, expiresAt: now.Add(ttl)}
	return nil
}

func (s *MemoryTicketStore) Consume(_ context.Context, hash string) (*Ticket, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	t, ok := s.tickets[hash]
	delete(s.tickets, hash)
	if !ok || time.Now().After(t.expiresAt) {
		return nil, nil
	}
	return &t.ticket, nil
}
