package ws

import (
	"encoding/json"
	"log"
	"sync"
)

type Hub struct {
	Rooms      map[string]map[*Client]bool
	Register   chan *Client
	Unregister chan *Client
	Inbound    chan *ClientMessage
	mu         sync.RWMutex
}

func NewHub() *Hub {
	return &Hub{
		Rooms:      make(map[string]map[*Client]bool),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Inbound:    make(chan *ClientMessage, 256),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.mu.Lock()
			if _, ok := h.Rooms[client.RoomID]; !ok {
				h.Rooms[client.RoomID] = make(map[*Client]bool)
			}
			h.Rooms[client.RoomID][client] = true
			h.mu.Unlock()
			log.Printf("ws: client %s joined room %s", client.PlayerID, client.RoomID)

		case client := <-h.Unregister:
			h.mu.Lock()
			if clients, ok := h.Rooms[client.RoomID]; ok {
				if _, exists := clients[client]; exists {
					delete(clients, client)
					close(client.Send)
					if len(clients) == 0 {
						delete(h.Rooms, client.RoomID)
					}
				}
			}
			h.mu.Unlock()
			log.Printf("ws: client %s left room %s", client.PlayerID, client.RoomID)

		case cm := <-h.Inbound:
			log.Printf("ws: message from %s in room %s: %s", cm.Client.PlayerID, cm.Client.RoomID, cm.Message.Type)
		}
	}
}

func (h *Hub) BroadcastToRoom(roomID string, msg Message) {
	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("ws: marshal error: %v", err)
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	if clients, ok := h.Rooms[roomID]; ok {
		for client := range clients {
			select {
			case client.Send <- data:
			default:
				close(client.Send)
				delete(clients, client)
			}
		}
	}
}

func (h *Hub) SendToPlayer(roomID, playerID string, msg Message) {
	data, err := json.Marshal(msg)
	if err != nil {
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	if clients, ok := h.Rooms[roomID]; ok {
		for client := range clients {
			if client.PlayerID == playerID {
				select {
				case client.Send <- data:
				default:
				}
			}
		}
	}
}
