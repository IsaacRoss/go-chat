// room
package main

import (
	"github.com/gorilla/websocket"
	"github.com/isaacross/trace"
	"log"
	"net/http"
)

type room struct {
	// forward is a channel that holds
	// incoming messages for other clients
	forward chan []byte
	join    chan *client
	leave   chan *client
	clients map[*client]bool

	tracer trace.Tracer
}

func (r *room) run() {
	for {
		select {
		case client := <-r.join:
			//joining
			r.clients[client] = true
			r.tracer.Trace("New client joined")
		case client := <-r.leave:
			delete(r.clients, client)
			close(client.send)
			r.tracer.Trace("Client Left")
		case msg := <-r.forward:
			r.tracer.Trace("Message Recieved: ", string(msg))
			//forward msg to all clients
			for client := range r.clients {
				select {
				case client.send <- msg:
					//send the msg
					r.tracer.Trace(" -- sent to client")
				default:
					//failed to send
					delete(r.clients, client)
					close(client.send)
					r.tracer.Trace(" -- failed to send, cleaned up client")
				}
			}
		}
	}
}

func newRoom() *room {
	return &room{
		forward: make(chan []byte),
		join:    make(chan *client),
		leave:   make(chan *client),
		clients: make(map[*client]bool),
	}
}

const (
	socketBufferSize  = 1024
	messageBufferSize = 256
)

var upgrader = &websocket.Upgrader{ReadBufferSize: socketBufferSize, WriteBufferSize: socketBufferSize}

func (r *room) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	socket, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		log.Fatal("ServeHTTP: ", err)
	}

	client := &client{
		socket: socket,
		send:   make(chan []byte, messageBufferSize),
		room:   r,
	}
	r.join <- client
	defer func() { r.leave <- client }()
	go client.write()
	client.read()
}
