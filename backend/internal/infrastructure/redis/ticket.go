package redis

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	goredis "github.com/redis/go-redis/v9"

	"github.com/newt239/chat/internal/usecase/realtime"
)

// TicketStore は WebSocket のチケットを全レプリカで共有します
type TicketStore struct {
	client *goredis.Client
}

func NewTicketStore(client *goredis.Client) *TicketStore {
	return &TicketStore{client: client}
}

func ticketKey(hash string) string {
	return "chat:ws:ticket:" + hash
}

func (s *TicketStore) Save(ctx context.Context, hash string, ticket realtime.Ticket, ttl time.Duration) error {
	payload, err := json.Marshal(ticket)
	if err != nil {
		return err
	}
	return s.client.Set(ctx, ticketKey(hash), payload, ttl).Err()
}

func (s *TicketStore) Consume(ctx context.Context, hash string) (*realtime.Ticket, error) {
	payload, err := s.client.GetDel(ctx, ticketKey(hash)).Bytes()
	if errors.Is(err, goredis.Nil) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var ticket realtime.Ticket
	if err := json.Unmarshal(payload, &ticket); err != nil {
		return nil, err
	}
	return &ticket, nil
}
