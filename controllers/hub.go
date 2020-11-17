// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package controllers

import (
	"log"
	"strconv"

	"websocket_test_1/models"
)

// GetHub returns the hub
func GetHub() *Hub {
	return &hub
}

// Run starts the hub
func (h *Hub) Run() {

	// Loop indefinitely
	for {

		//If any of these communication received (through these channels)
		select {

		// Register request from a client
		case client := <-h.register:

			// Add client to the registered clients map of hub

			//Get the clients which are already registerted for that userProfileId
			existingClients := h.clients[client.userProfileID]

			//If no client is already registered for that userProfileId, create an empty set to hold the client
			if existingClients == nil {
				existingClients = map[*Client]bool{}
			}

			//Add the current client to the set of registerted clients
			existingClients[client] = true

			//Assign the set to that userProfileId
			h.clients[client.userProfileID] = existingClients

		// Unregister request from a client
		case client := <-h.unregister:

			// If the current client exists in registered client map
			if existingClients, ok := h.clients[client.userProfileID]; ok {
				if _, ok := existingClients[client]; ok {

					// Delete the client from registered client map and close the channel
					delete(existingClients, client)
					close(client.send)
				}
			}

		// Inbound message from a client
		case message := <-h.broadcast:

			//Forward to all the registered clients of sender, for syncing

			// Range over all the currently registed clients for the sender
			// Also, might use a separate goroutine to broadcast every time a new message comes, instead of using a single
			// goroutine to broadcast all the messages in the application, which might potentially slow the message
			// delivery down, in case of increased traffic, as broadcasting (using loop) is blocking.
			log.Println("Forwarding message: " + message.Content +
				" to " + strconv.Itoa(len(h.clients[message.SenderID])) + " sender clients with userProfileId:" + message.SenderID)
			for client := range h.clients[message.SenderID] {
				select {

				// Forward the message to the client
				case client.send <- message:

				// If the client is not accepting the message
				default:

					// Delete the client and close the channel
					delete(h.clients[message.SenderID], client)
					close(client.send)
				}
			}

			//Forward to all the registered clients of recipient

			// Range over all the currently registed clients for the recipient
			// Also, might use a separate goroutine to broadcast every time a new message comes, instead of using a single
			// goroutine to broadcast all the messages in the application, which might potentially slow the message
			// delivery down, in case of increased traffic, as broadcasting (using loop) is blocking.
			log.Println("Forwarding message: " + message.Content +
				" to " + strconv.Itoa(len(h.clients[message.RecipientID])) + " recipient clients with userProfileId:" + message.RecipientID)
			for client := range h.clients[message.RecipientID] {
				select {

				// Forward the message to the client
				case client.send <- message:

				// If the client is not accepting the message
				default:

					// Delete the client and close the channel
					delete(h.clients[message.RecipientID], client)
					close(client.send)
				}
			}
		}
	}
}

// Hub maintains the set of active clients and broadcasts messages to them
type Hub struct {

	// Registered clients, stored in userProfileId : Set<Client> format.
	clients map[string](map[*Client]bool)

	// Channel to receive inbound messages from a client
	broadcast chan models.Message

	// Channel to receive register request from a client
	register chan *Client

	// Channel to receive unregister request from a client
	unregister chan *Client
}

// Maintaining a single "instance" of hub
var hub = Hub{
	broadcast:  make(chan models.Message),
	register:   make(chan *Client),
	unregister: make(chan *Client),
	clients:    make(map[string](map[*Client]bool)),
}
