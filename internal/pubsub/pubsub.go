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

// ------- adapter to provide simple string-channel Subscriber for ws package -------
type redisSubscriberAdapter struct {
	pubsub *redis.PubSub
	ch     chan string
	done   chan struct{}
}

func (a *redisSubscriberAdapter) Channel() <-chan string {
	return a.ch
}

func (a *redisSubscriberAdapter) Close() error {
	close(a.done)
	_ = a.pubsub.Close()
	close(a.ch)
	return nil
}

// SubscribeAdapter returns a Subscriber that yields string payloads for the given channel.
func (r *RedisPubSub) SubscribeAdapter(ctx context.Context, channel string) (interface{ Channel() <-chan string; Close() error }, error) {
	ps := r.client.Subscribe(ctx, channel)
	// create adapter
	adapter := &redisSubscriberAdapter{
		pubsub: ps,
		ch:     make(chan string, 16),
		done:   make(chan struct{}),
	}
	// forward messages
	go func() {
		ch := ps.Channel()
		for {
			select {
			case <-adapter.done:
				return
			case msg, ok := <-ch:
				if !ok {
					return
				}
				adapter.ch <- msg.Payload
			}
		}
	}()
	return adapter, nil
}
