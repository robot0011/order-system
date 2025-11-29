package main

import (
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/websocket"
)

func main() {
	// Use the token from the error message
	token := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NjQ0MTIyNzUsInVzZXJuYW1lIjoiZ3VndWdhZ2EifQ.TlL-i39CM5sCc8XuO8VvwYtd1zpdX82KvPC-0PXQllY"
	
	// Connect to WebSocket server
	url := "ws://localhost:3000/ws/orders?token=" + token
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		log.Fatal("WebSocket connection error:", err)
	}
	defer conn.Close()

	log.Println("Connected to WebSocket server")

	// Wait for interrupt signal to gracefully close the connection
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	
	done := make(chan bool, 1)
	go func() {
		defer close(done)
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				log.Println("Read error:", err)
				return
			}
			log.Printf("Received: %s", message)
		}
	}()

	select {
	case <-done:
		log.Println("Connection closed")
	case <-c:
		log.Println("Interrupt received, closing connection")
		// Cleanly close the connection by sending a close message
		err := conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		if err != nil {
			log.Println("Write close error:", err)
			return
		}
		select {
		case <-done:
		case <-time.After(time.Second):
		}
	}
}