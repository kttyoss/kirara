package main

import (
	"log"
	"sync"
	"time"

	"github.com/goccy/go-json"
	"github.com/gofiber/websocket/v2"
)

type Kclient struct {
	identifier string
	registered bool
	isClosing  bool
	mu         sync.Mutex
}

type Kmessage struct {
	connection *websocket.Conn
	message    interface{}
}

var clients = make(map[*websocket.Conn]*Kclient)
var register = make(chan *websocket.Conn)
var unregister = make(chan *websocket.Conn)
var broadcast = make(chan *Kmessage)

func setIdentifier(conn *websocket.Conn, identifier string) {
	clients[conn].identifier = identifier
	clients[conn].registered = true
	log.Printf("registered %s as %s", conn.RemoteAddr().String(), identifier)
}

func runHub() {
	for {
		select {
		case c := <-register:
			clients[c] = &Kclient{identifier: "UNIDENT", registered: false}
		case c := <-unregister:
			delete(clients, c)
		case m := <-broadcast:
			if !clients[m.connection].registered {
				m.connection.WriteMessage(websocket.TextMessage, []byte(`{"event": "error", "message": "Your node isn't registered. Please send a registration packet before sending any data.", "code": "ERR_NOT_REGISTERED", "success": false}`))
				return
			}
			for conn, c := range clients {
				go func(conn *websocket.Conn, c *Kclient) {
					if conn == m.connection {
						return
					}
					c.mu.Lock()
					defer c.mu.Unlock()
					if c.isClosing || !c.registered {
						return
					}
					data := make(map[string]interface{})

					data["event"] = "data"
					data["node"] = clients[m.connection].identifier
					data["data"] = m.message
					data["timestamp"] = time.Now().UnixMilli()
					log.Println(data)
					dataBytes, err := json.Marshal(data)
					log.Println(string(dataBytes))
					if err != nil {
						log.Printf("error: %v", err)
						return
					}
					if err := conn.WriteMessage(websocket.TextMessage, dataBytes); err != nil {
						c.isClosing = true

						conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseInternalServerErr, "internal server error"))
						conn.Close()
						unregister <- conn
					}
				}(conn, c)
			}
		}
	}
}
