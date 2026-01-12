package pubsub

import (
	"context"
	"encoding/json"
	"log"

	"github.com/redis/go-redis/v9"
)

const (
	ChannelRoomMessage = "room:message"
	ChannelUserMessage = "user:message"
	ChannelPresence    = "presence"
)

type Message struct {
	Type    string          `json:"type"`
	RoomID  uint64          `json:"room_id,omitempty"`
	UserID  uint64          `json:"user_id,omitempty"`
	Payload json.RawMessage `json:"payload"`
}

type RedisPubSub struct {
	client    *redis.Client
	pubsub    *redis.PubSub
	handlers  map[string]func(*Message)
}

func NewRedisPubSub(client *redis.Client) *RedisPubSub {
	return &RedisPubSub{
		client:   client,
		handlers: make(map[string]func(*Message)),
	}
}

func (r *RedisPubSub) Subscribe(ctx context.Context, channels ...string) error {
	r.pubsub = r.client.Subscribe(ctx, channels...)

	_, err := r.pubsub.Receive(ctx)
	if err != nil {
		return err
	}

	go r.listen(ctx)
	return nil
}

func (r *RedisPubSub) listen(ctx context.Context) {
	ch := r.pubsub.Channel()

	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-ch:
			if !ok {
				return
			}
			r.handleMessage(msg)
		}
	}
}

func (r *RedisPubSub) handleMessage(msg *redis.Message) {
	var m Message
	if err := json.Unmarshal([]byte(msg.Payload), &m); err != nil {
		log.Printf("Failed to unmarshal pubsub message: %v", err)
		return
	}

	if handler, ok := r.handlers[msg.Channel]; ok {
		handler(&m)
	}
}

func (r *RedisPubSub) OnMessage(channel string, handler func(*Message)) {
	r.handlers[channel] = handler
}

func (r *RedisPubSub) Publish(ctx context.Context, channel string, msg *Message) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	return r.client.Publish(ctx, channel, data).Err()
}

func (r *RedisPubSub) PublishRoomMessage(ctx context.Context, roomID uint64, payload []byte) error {
	msg := &Message{
		Type:    "room_message",
		RoomID:  roomID,
		Payload: payload,
	}
	return r.Publish(ctx, ChannelRoomMessage, msg)
}

func (r *RedisPubSub) PublishUserMessage(ctx context.Context, userID uint64, payload []byte) error {
	msg := &Message{
		Type:    "user_message",
		UserID:  userID,
		Payload: payload,
	}
	return r.Publish(ctx, ChannelUserMessage, msg)
}

func (r *RedisPubSub) PublishPresence(ctx context.Context, userID uint64, status string) error {
	payload, _ := json.Marshal(map[string]interface{}{
		"user_id": userID,
		"status":  status,
	})
	msg := &Message{
		Type:    "presence",
		UserID:  userID,
		Payload: payload,
	}
	return r.Publish(ctx, ChannelPresence, msg)
}

func (r *RedisPubSub) Close() error {
	if r.pubsub != nil {
		return r.pubsub.Close()
	}
	return nil
}