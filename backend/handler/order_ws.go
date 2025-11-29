package handler

import (
	"encoding/json"
	"log"
	"order-system/database"
	"order-system/models"
	"sync"

	"github.com/gofiber/websocket/v2"
)

type wsClient struct {
	conn          *websocket.Conn
	send          chan []byte
	restaurantIDs map[uint]struct{}
}

type OrderEvent struct {
	Type  string        `json:"type"`
	Order OrderResponse `json:"order"`
}

type orderHub struct {
	clients    map[*wsClient]struct{}
	broadcast  chan OrderEvent
	register   chan *wsClient
	unregister chan *wsClient
	mu         sync.Mutex
}

var globalOrderHub = newOrderHub()

func init() {
	go globalOrderHub.run()
}

func newOrderHub() *orderHub {
	return &orderHub{
		clients:    map[*wsClient]struct{}{},
		broadcast:  make(chan OrderEvent, 32),
		register:   make(chan *wsClient),
		unregister: make(chan *wsClient),
	}
}

func (h *orderHub) run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = struct{}{}
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
		case event := <-h.broadcast:
			payload, err := json.Marshal(event)
			if err != nil {
				log.Println("failed to marshal order event:", err)
				continue
			}
			for client := range h.clients {
				if len(client.restaurantIDs) == 0 {
					continue
				}
				if _, ok := client.restaurantIDs[event.Order.RestaurantID]; !ok {
					continue
				}
				select {
				case client.send <- payload:
				default:
					delete(h.clients, client)
					close(client.send)
				}
			}
		}
	}
}

func (h *orderHub) publish(eventType string, order OrderResponse) {
	h.broadcast <- OrderEvent{
		Type:  eventType,
		Order: order,
	}
}

func HandleOrderSocket(c *websocket.Conn) {
	token := c.Query("token")
	log.Printf("WebSocket connection attempt with token: %s", token)
	username, err := ValidateAccessToken(token)
	if err != nil {
		log.Printf("WebSocket authentication failed: %v", err)
		c.WriteMessage(websocket.TextMessage, []byte("invalid token"))
		c.Close()
		return
	}
	log.Printf("WebSocket authenticated user: %s", username)

	restaurantIDs, err := fetchRestaurantIDs(username)
	if err != nil {
		log.Printf("WebSocket unable to load restaurant subscriptions for user %s: %v", username, err)
		c.WriteMessage(websocket.TextMessage, []byte("unable to load restaurant subscriptions"))
		c.Close()
		return
	}

	client := newWSClient(c, restaurantIDs)
	globalOrderHub.register <- client

	defer func() {
		globalOrderHub.unregister <- client
		c.Close()
	}()

	go client.writePump()
	client.readPump()
}

func newWSClient(conn *websocket.Conn, restaurantIDs []uint) *wsClient {
	idSet := make(map[uint]struct{}, len(restaurantIDs))
	for _, id := range restaurantIDs {
		idSet[id] = struct{}{}
	}
	return &wsClient{
		conn:          conn,
		send:          make(chan []byte, 32),
		restaurantIDs: idSet,
	}
}

func fetchRestaurantIDs(username string) ([]uint, error) {
	var user models.User
	if err := database.DB.Preload("Restaurants").Where("username = ?", username).First(&user).Error; err != nil {
		return nil, err
	}

	ids := make([]uint, 0, len(user.Restaurants))
	for _, restaurant := range user.Restaurants {
		ids = append(ids, restaurant.ID)
	}
	return ids, nil
}

func (c *wsClient) readPump() {
	for {
		if _, _, err := c.conn.ReadMessage(); err != nil {
			break
		}
	}
}

func (c *wsClient) writePump() {
	for message := range c.send {
		if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
			break
		}
	}
}
