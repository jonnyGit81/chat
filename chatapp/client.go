package main

import (
	"fmt"
	"github.com/gorilla/websocket"
	"time"
)

// client represents a single chatting user.
type client struct {
	// socket is the web socket for this client.
	socket *websocket.Conn
	// send is a channel on which messages are sent.

	//we change to message struct
	//send chan []byte
	send chan *message

	// room is the room this client is chatting in.
	room *room
	// userdata
	userData map[string]interface{}
}

// The read method allows our client to read from the socket via the ReadMessage method,
// continually sending any received messages to the forward channel on the room type.
// If it encounters an error (such as 'the socket has died'),
// the loop will break and the socket will be closed.
func (c *client) read() {
	for {
		var msg *message
		err := c.socket.ReadJSON(&msg)
		if err != nil {
			return
		}
		msg.When = time.Now()
		msg.Name = c.userData["name"].(string)

		// Instead of get directly from client we created a Avatar Interface
		// if avatarUrl, ok := c.userData["avatar_url"] ; ok {
		//	msg.AvatarURL = avatarUrl.(string)
		// }

		// room have avatar, we passing in nil struct, go allowed to call method from nil object,
		// then we passing in avatar.GetAvatarURL(c), which is extract avatar url from client.
		// so it make us easy to make a test case right? isn't it?
		// cool stuff. also we can have another implementation to get from Gravatar instead of user gmail avatar.
		// NICE
		msg.AvatarURL, _ = c.room.avatar.GetAvatarURL(c)

		fmt.Println("c", *c)

		c.room.forward <- msg

		//if _, msg, err := c.socket.ReadMessage(); err == nil {
		//	c.room.forward <- msg
		//	fmt.Println("User type message, c.room.forward <- msg ", string(msg[:]))
		//}else {
		//	break
		//}

	}
	c.socket.Close()
}

// the write method continually accepts messages from the send channel writing everything out
// of the socket via the WriteMessage method. If writing to the socket fails,
// the for loop is broken and the socket is closed
func (c *client) write() {
	for msg := range c.send { // equal to msg byte[] <- ch, waiting for receive
		err := c.socket.WriteJSON(msg)

		if err != nil {
			break
		}

		//if err := c.socket.WriteMessage(websocket.TextMessage, msg); err != nil {
		//	fmt.Println("write error", err)
		//	break;
		//}
	}
	c.socket.Close()
}
