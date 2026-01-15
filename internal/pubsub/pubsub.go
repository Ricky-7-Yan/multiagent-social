package pubsub

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/redis/go-redis/v9"
)

// RedisPubSub is a minimal wrapper for publish/subscribe.
type RedisPubSub struct {
	client *redis.Client
}

// NewRedisPubSub connects to redis at addr (host:port).
func NewRedisPubSub(addr string) (*RedisPubSub, error) {
	opt := &redis.Options{
		Addr: addr,
	}
	c := redis.NewClient(opt)
	if err := c.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("redis ping: %w", err)
	}
	return &RedisPubSub{client: c}, nil
}

// Publish sends a JSON-encoded message on the channel.
func (r *RedisPubSub) Publish(ctx context.Context, channel string, v interface{}) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return r.client.Publish(ctx, channel, b).Err()
}

// Subscribe returns a go-redis PubSub for the given channel.
func (r *RedisPubSub) Subscribe(ctx context.Context, channel string) *redis.PubSub {
	return r.client.Subscribe(ctx, channel)
}

// Client returns the underlying redis client for advanced usage.
func (r *RedisPubSub) Client() *redis.Client {
	return r.client
}

