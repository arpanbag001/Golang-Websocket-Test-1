// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"

	"websocket_test_1/main/models"
)

// handleWebsocket handles websocket requests
func handleWebsocket(w http.ResponseWriter, r *http.Request) {

	// Check origin of the request. For now, allowing every request by returning true without checking
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }

	// Upgrade the request to websocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	log.Println("Client connected!")

	// Register the client
	client := &Client{conn: conn, send: make(chan models.Message)}
	getHub().register <- client

	// Read and write messages in new goroutines
	go client.writePump()
	go client.readPump()
}

// readPump pumps messages from the websocket connection to the hub.
// The application runs readPump in a per-connection goroutine. The application
// ensures that there is at most one reader on a connection by executing all
// reads from this goroutine.
func (c *Client) readPump() {

	// Defer unregistering client and closing the connection
	defer func() {
		getHub().unregister <- c
		c.conn.Close()
	}()

	// Configure
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	// Loop indefinitely
	for {

		// Read the message
		message := models.Message{}
		err := c.conn.ReadJSON(&message)

		// If got an error while reading the message
		if err != nil {

			// If error does not indicate client is disconnecting normally
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNoStatusReceived) {
				log.Println(err)
			}

			log.Println("Client disconnected!")
			break
		}

		// Broadcast the received message to the hub, to be sent to all the recipients
		getHub().broadcast <- message
	}
}

// writePump pumps messages from the hub to the websocket connection.
// A goroutine running writePump is started for each connection. The
// application ensures that there is at most one writer to a connection by
// executing all writes from this goroutine.
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)

	// Defer closing the connection
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	// Loop indefinitely
	for {

		// If any of these communication received (through these channels)
		select {

		// Message received
		case message, ok := <-c.send:

			c.conn.SetWriteDeadline(time.Now().Add(writeWait))

			// If hub closed the channel
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			// Send (write) the message
			if err := c.conn.WriteJSON(message); err != nil {
				return
			}

		// Ticker ticked
		case <-ticker.C:

			c.conn.SetWriteDeadline(time.Now().Add(writeWait))

			// Do ping
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// Client represents a client connected to the websocket server
type Client struct {

	// The websocket connection
	conn *websocket.Conn

	// Buffered channel of outbound messages
	send chan models.Message
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

const (
	// Time allowed to write a message to the client
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the client
	pongWait = 60 * time.Second

	// Send pings to client with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from client
	maxMessageSize = 512
)
