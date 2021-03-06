package main

import (
	"log"
	"net/http"

	"github.com/coffemanfp/trace"
	"github.com/gorilla/websocket"
	"github.com/stretchr/objx"
)

const (
	socketBufferSize  = 1024
	messageBufferSize = 256
)

var upgrader = &websocket.Upgrader{
	ReadBufferSize:  socketBufferSize,
	WriteBufferSize: socketBufferSize,
}

type room struct {
	// forward is a channel that holds incoming messages
	// that should be forwarded to the other clients
	forward chan *message

	// join is a channel for clients wishing to join the room.
	join chan *client

	// leave is a channel for clients wishing to leave the room.
	leave chan *client

	// clients holds all current clients in this room.
	clients map[*client]bool

	// tracer will receive trace information of activity in the room.
	tracer trace.Tracer
}

func newRoom() *room {
	return &room{
		forward: make(chan *message),
		join:    make(chan *client),
		leave:   make(chan *client),
		clients: make(map[*client]bool),
		tracer:  trace.Off(),
	}
}

func (rm *room) run() {
	for {
		select {
		case client := <-rm.join:
			// joining
			rm.clients[client] = true
			rm.tracer.Trace("New client joined")
		case client := <-rm.leave:
			// leaving
			delete(rm.clients, client)
			close(client.send)
			rm.tracer.Trace("Client left")
		case msg := <-rm.forward:
			for client := range rm.clients {
				select {
				case client.send <- msg:
					// send the message
					rm.tracer.Trace(" -- sent to client")
				default:
					// failed to send
					delete(rm.clients, client)
					close(client.send)
					rm.tracer.Trace(" -- failed to send, cleaned up client")
				}
			}
		}
	}
}

// ServeHTTP A ServeHTTP implementation.
func (rm *room) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	socket, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal("ServeHTTP:", err)
		return
	}

	authCookie, err := r.Cookie("auth")
	if err != nil {
		log.Fatal("Failed to get auth cookie:", err)
		return
	}

	client := &client{
		socket:   socket,
		send:     make(chan *message, messageBufferSize),
		room:     rm,
		userData: objx.MustFromBase64(authCookie.Value),
	}

	rm.join <- client
	defer func() { rm.leave <- client }()
	go client.write()
	client.read()
}
