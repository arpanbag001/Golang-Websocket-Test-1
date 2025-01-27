// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package controllers

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/julienschmidt/httprouter"

	"websocket_test_1/database/dao"
	"websocket_test_1/models"
	"websocket_test_1/utils/config"
)

// HandleWebsocket handles websocket requests
func HandleWebsocket(w http.ResponseWriter, r *http.Request) {

	//Get the userProfileId from request params
	userProfileID := httprouter.ParamsFromContext(r.Context()).ByName("userProfileId")

	// Check origin of the request. For now, allowing every request by returning true without checking
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }

	// Upgrade the request to websocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	log.Println("Client with userProfileId: " + userProfileID + " connected!")

	// Register the client
	client := &Client{userProfileID: userProfileID, conn: conn, send: make(chan models.Message)}
	GetHub().register <- client

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
		GetHub().unregister <- c
		c.conn.Close()
	}()

	// Configure
	c.conn.SetReadLimit(config.MaxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(config.PongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(config.PongWait)); return nil })

	// Loop indefinitely
	for {

		//Acceptable request JSON body
		type MessageRequestFormat struct {
			RecipientID string `json:"recipientId"`
			Content     string `json:"content"`
		}

		//Get the message details provided as JSON
		var messageDetails MessageRequestFormat
		err := c.conn.ReadJSON(&messageDetails)

		// If got an error while reading the message
		if err != nil {

			// If error does not indicate client is disconnecting normally
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNoStatusReceived) {
				log.Println(err)
			}

			log.Println("Client with userProfileId: " + c.userProfileID + " disconnected!")
			break
		}

		//Get additional details of the message.
		//TODO: Check whether the sender has permission to send message to recipient

		//Create and store the message
		createdMessage, err := dao.CreateMessage(c.userProfileID, messageDetails.RecipientID, messageDetails.Content)

		if err != nil {
			log.Println("Client with userProfileId: " + c.userProfileID + " disconnected!")
			break
		}

		// Broadcast the received message to the hub, to be sent to all the recipients
		GetHub().broadcast <- *createdMessage
	}
}

// writePump pumps messages from the hub to the websocket connection.
// A goroutine running writePump is started for each connection. The
// application ensures that there is at most one writer to a connection by
// executing all writes from this goroutine.
func (c *Client) writePump() {
	ticker := time.NewTicker(config.PingPeriod)

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

			c.conn.SetWriteDeadline(time.Now().Add(config.WriteWait))

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

			c.conn.SetWriteDeadline(time.Now().Add(config.WriteWait))

			// Do ping
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// Client represents a client connected to the websocket server
type Client struct {

	// The userProfileId
	userProfileID string

	// The websocket connection
	conn *websocket.Conn

	// Buffered channel of outbound messages
	send chan models.Message
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}
