package main

import (
	"github.com/gorilla/websocket"
	"time"
)

// client represents a single chatting user
type client struct {
	// this is ws for this client
	socket *websocket.Conn
	// channel to send messags
	send chan *message

	room *room

	userData map[string]interface{}
}

func (c *client) read() {
	for {
		var msg *message
		if err := c.socket.ReadJSON(&msg); err == nil {
			msg.When = time.Now()
			msg.Name = c.userData["name"].(string)
			c.room.forward <- msg
		} else {
			break
		}
	}
	c.socket.Close()
}
func (c *client) write() {
	for msg := range c.send {
		if err := c.socket.WriteJSON(msg); err != nil {
			break
		}
	}
	c.socket.Close()
}
