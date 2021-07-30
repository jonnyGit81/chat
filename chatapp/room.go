package main

import (
	"github.com/gorilla/websocket"
	"github.com/jonnyGit81/chat/chatapp/trace"
	"github.com/stretchr/objx"
	"log"
	"net/http"
)

const (
	socketBufferSize  = 1024
	messageBufferSize = 256
)

// to upgrade HTTP Connection as Web Socket
var upgrader = &websocket.Upgrader{ReadBufferSize: socketBufferSize, WriteBufferSize: socketBufferSize}

// We need a way for clients to join and leave rooms in order to ensure that the c.room.forward <- msg code
// in the preceding section actually forwards the message to all the clients.
// To ensure that we are not trying to access the same data at the same time,
// a sensible approach is to use two channels: one that will add a client to the room and another
// that will remove it.
type room struct {
	// forward is a channel that holds incoming messages
	// that should be forwarded to the other clients.
	//forward chan []byte

	//change to message struct
	forward chan *message

	// join is a channel for clients wishing to join the room
	join chan *client

	// leave is a channel for clients wishing to leave the room
	leave chan *client

	// clients holds all current clients in this room.
	// If we were to access the map directly,
	// it is possible that two Go routines running concurrently might try to modify the map at the same time
	// resulting in corrupt memory or an unpredictable state.
	clients map[*client]bool

	//for loging trace
	tracer trace.Tracer

	// avatar is how avatar information will be obtained.
	avatar Avatar
}

// Concurrency programming using idiomatic Go
// Now we get to use an extremely powerful feature of Go's concurrency offerings—the select statement.
// We can use select statements whenever we need to synchronize or modify shared memory,
// or take different actions depending on the various activities within our channels.
/*
The preceding code will keep watching the three channels inside our room: join,
leave, and forward. If a message is received on any of those channels,
the select statement will run the code for that particular case.
It is important to remember that it will only run one block of case code at a time.
This is how we are able to synchronize to ensure that our r.clients map is only ever modified by one thing at a time.
*/
func (r *room) run() {
	for {
		select {
		case clientJoin := <-r.join:
			//joining room
			r.clients[clientJoin] = true
			r.tracer.Trace("User joining room <-r.join", clientJoin.userData["name"].(string))
			//fmt.Println("User joining room <-r.join")
		case clientLeave := <-r.leave:
			//leaving room
			delete(r.clients, clientLeave)
			close(clientLeave.send)
			r.tracer.Trace("User leave room <-r.leave")
			//fmt.Println("User leave room <-r.leave")
		case msg := <-r.forward:
			// forward message to all clients
			for c := range r.clients {
				select {
				case c.send <- msg:
					// send message to client
					r.tracer.Trace("Message received: ", msg.Message)
				default:
					// failed to send, remove the client and close the client chanel
					// for prevent blocking go routine that receiver never pickup
					delete(r.clients, c)
					close(c.send)
					r.tracer.Trace("User sending message c.send <-msg Failed, going to close channel", c.send)
					//fmt.Println("User sending message c.send <-msg Failed, going to close channel", c.send)
				}
			}
		}
	}
}

// Factory function
// Update the newRoom function so that we can pass in an Avatar implementation for use;
// we will just assign this implementation to the new field when we create our room instance:
//func newRoom() *room {
func newRoom(avatar Avatar) *room {
	r := room{
		//we change to message struct
		//forward: make(chan []byte),
		forward: make(chan *message),
		join:    make(chan *client),
		leave:   make(chan *client),
		clients: make(map[*client]bool),
		tracer:  trace.Off(),
		avatar:  avatar,
	}
	return &r
}

// Turning a room into an HTTP handler
// In order to use web sockets, we must upgrade the HTTP connection using the websocket.Upgrader type,
// which is reusable so we need only create one. Then, when a request comes in via the ServeHTTP method,
// we get the socket by calling the upgrader.Upgrade method.
// All being well, we then create our client and pass it into the join channel for the current room.
// We also defer the leaving operation for when the client is finished,
// which will ensure everything is tidied up after a user goes away.
func (r *room) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	socket, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		log.Fatal("ServeHTTP:", err)
		return
	}

	authCookie, err := req.Cookie("auth")
	if err != nil {
		log.Fatal("Failed to get auth cookie:", err)
		return
	}

	client := &client{
		socket:   socket,
		send:     make(chan *message, messageBufferSize),
		room:     r,
		userData: objx.MustFromBase64(authCookie.Value),
	}

	r.join <- client

	defer func() { r.leave <- client }()

	// The write method for the client is then called as a Go routine
	go client.write()

	// Finally, we call the read method in the main thread,
	// which will block operations (keeping the connection alive) until it's time to close it
	client.read()
}
