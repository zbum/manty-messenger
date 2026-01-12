package websocket

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	"Mmessenger/internal/pubsub"
)

type Hub struct {
	clients    map[*Client]bool
	rooms      map[uint64]map[*Client]bool
	userConns  map[uint64]*Client
	broadcast  chan *BroadcastMessage
	register   chan *Client
	unregister chan *Client
	mu         sync.RWMutex
	pubsub     *pubsub.RedisPubSub
}

type BroadcastMessage struct {
	RoomID  uint64
	Message []byte
	Sender  *Client
}

func NewHub(ps *pubsub.RedisPubSub) *Hub {
	h := &Hub{
		clients:    make(map[*Client]bool),
		rooms:      make(map[uint64]map[*Client]bool),
		userConns:  make(map[uint64]*Client),
		broadcast:  make(chan *BroadcastMessage, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		pubsub:     ps,
	}

	if ps != nil {
		h.setupPubSubHandlers()
	}

	return h
}

func (h *Hub) setupPubSubHandlers() {
	h.pubsub.OnMessage(pubsub.ChannelRoomMessage, func(msg *pubsub.Message) {
		h.handlePubSubRoomMessage(msg)
	})

	h.pubsub.OnMessage(pubsub.ChannelUserMessage, func(msg *pubsub.Message) {
		h.handlePubSubUserMessage(msg)
	})

	h.pubsub.OnMessage(pubsub.ChannelPresence, func(msg *pubsub.Message) {
		h.handlePubSubPresence(msg)
	})
}

func (h *Hub) handlePubSubRoomMessage(msg *pubsub.Message) {
	h.mu.RLock()
	room, ok := h.rooms[msg.RoomID]
	h.mu.RUnlock()

	if !ok {
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	for client := range room {
		select {
		case client.send <- msg.Payload:
		default:
		}
	}
}

func (h *Hub) handlePubSubUserMessage(msg *pubsub.Message) {
	h.mu.RLock()
	client, ok := h.userConns[msg.UserID]
	h.mu.RUnlock()

	if ok {
		select {
		case client.send <- msg.Payload:
		default:
		}
	}
}

func (h *Hub) handlePubSubPresence(msg *pubsub.Message) {
	var payload struct {
		UserID uint64 `json:"user_id"`
		Status string `json:"status"`
	}
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		log.Printf("Failed to unmarshal presence payload: %v", err)
		return
	}

	wsMsg := &WSMessage{
		Type: TypePresenceUpdate,
		Payload: map[string]interface{}{
			"user_id": payload.UserID,
			"status":  payload.Status,
		},
		Timestamp: time.Now(),
	}

	data, err := marshalMessage(wsMsg)
	if err != nil {
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	for client := range h.clients {
		if client.UserID != payload.UserID {
			select {
			case client.send <- data:
			default:
			}
		}
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.userConns[client.UserID] = client
			h.mu.Unlock()

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				delete(h.userConns, client.UserID)
				close(client.send)

				// Remove from all rooms
				for roomID := range client.rooms {
					if room, ok := h.rooms[roomID]; ok {
						delete(room, client)
						if len(room) == 0 {
							delete(h.rooms, roomID)
						}
					}
				}
			}
			h.mu.Unlock()

		case msg := <-h.broadcast:
			h.mu.RLock()
			if room, ok := h.rooms[msg.RoomID]; ok {
				for client := range room {
					if client != msg.Sender {
						select {
						case client.send <- msg.Message:
						default:
							h.mu.RUnlock()
							h.mu.Lock()
							close(client.send)
							delete(h.clients, client)
							delete(room, client)
							h.mu.Unlock()
							h.mu.RLock()
						}
					}
				}
			}
			h.mu.RUnlock()
		}
	}
}

func (h *Hub) JoinRoom(client *Client, roomID uint64) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.rooms[roomID] == nil {
		h.rooms[roomID] = make(map[*Client]bool)
	}
	h.rooms[roomID][client] = true
	client.rooms[roomID] = true
}

func (h *Hub) LeaveRoom(client *Client, roomID uint64) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if room, ok := h.rooms[roomID]; ok {
		delete(room, client)
		if len(room) == 0 {
			delete(h.rooms, roomID)
		}
	}
	delete(client.rooms, roomID)
}

func (h *Hub) BroadcastToRoom(roomID uint64, message []byte, sender *Client) {
	// Send to local clients
	h.broadcast <- &BroadcastMessage{
		RoomID:  roomID,
		Message: message,
		Sender:  sender,
	}

	// Publish to Redis for other servers
	if h.pubsub != nil {
		if err := h.pubsub.PublishRoomMessage(context.Background(), roomID, message); err != nil {
			log.Printf("Failed to publish room message to Redis: %v", err)
		}
	}
}

func (h *Hub) SendToUser(userID uint64, message []byte) {
	h.mu.RLock()
	client, ok := h.userConns[userID]
	h.mu.RUnlock()

	if ok {
		select {
		case client.send <- message:
		default:
		}
	}

	// Also publish to Redis for other servers
	if h.pubsub != nil {
		if err := h.pubsub.PublishUserMessage(context.Background(), userID, message); err != nil {
			log.Printf("Failed to publish user message to Redis: %v", err)
		}
	}
}

func (h *Hub) GetRoomMembers(roomID uint64) []uint64 {
	h.mu.RLock()
	defer h.mu.RUnlock()

	var userIDs []uint64
	if room, ok := h.rooms[roomID]; ok {
		for client := range room {
			userIDs = append(userIDs, client.UserID)
		}
	}
	return userIDs
}

func (h *Hub) IsUserOnline(userID uint64) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	_, ok := h.userConns[userID]
	return ok
}

func (h *Hub) BroadcastPresence(userID uint64, status string) {
	msg := &WSMessage{
		Type: TypePresenceUpdate,
		Payload: map[string]interface{}{
			"user_id": userID,
			"status":  status,
		},
		Timestamp: time.Now(),
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	for client := range h.clients {
		if client.UserID != userID {
			if data, err := marshalMessage(msg); err == nil {
				select {
				case client.send <- data:
				default:
				}
			}
		}
	}

	// Publish to Redis for other servers
	if h.pubsub != nil {
		if err := h.pubsub.PublishPresence(context.Background(), userID, status); err != nil {
			log.Printf("Failed to publish presence to Redis: %v", err)
		}
	}
}

// RoomInfo for invite payload
type RoomResponseInfo interface {
	GetID() uint64
	GetName() string
	GetDescription() *string
	GetRoomType() string
	GetMemberCount() int
}

func (h *Hub) SendRoomInvite(userID uint64, room interface {
	GetID() uint64
	GetName() string
	GetDescription() *string
	GetRoomType() string
	GetMemberCount() int
}) {
	payload := &RoomInvitedPayload{
		Room: &RoomInfo{
			ID:          room.GetID(),
			Name:        room.GetName(),
			Description: room.GetDescription(),
			RoomType:    room.GetRoomType(),
			MemberCount: room.GetMemberCount(),
		},
	}

	msg := &WSMessage{
		Type:      TypeRoomInvited,
		Payload:   payload,
		Timestamp: time.Now(),
	}

	if data, err := marshalMessage(msg); err == nil {
		h.SendToUser(userID, data)
	}
}
