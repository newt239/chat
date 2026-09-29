package redis

import (
	"context"

	goredis "github.com/redis/go-redis/v9"
)

const eventsChannel = "chat:ws:events"

// Broker は WebSocket のイベントを Redis の Pub/Sub で全レプリカに配ります
type Broker struct {
	client *goredis.Client
}

func NewBroker(client *goredis.Client) *Broker {
	return &Broker{client: client}
}

func (b *Broker) Publish(ctx context.Context, payload []byte) error {
	return b.client.Publish(ctx, eventsChannel, payload).Err()
}

// Subscribe は ctx が終わるまで受信したイベントを handle に渡します。切断時は go-redis が再接続します
func (b *Broker) Subscribe(ctx context.Context, handle func([]byte)) error {
	sub := b.client.Subscribe(ctx, eventsChannel)
	defer func() { _ = sub.Close() }()
	if _, err := sub.Receive(ctx); err != nil {
		return err
	}
	messages := sub.Channel()
	for {
		select {
		case <-ctx.Done():
			return nil
		case msg, ok := <-messages:
			if !ok {
				return nil
			}
			handle([]byte(msg.Payload))
		}
	}
}
