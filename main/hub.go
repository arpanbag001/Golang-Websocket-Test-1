// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import "websocket_test_1/main/models"

// Hub maintains the set of active clients and broadcasts messages to them
type Hub struct {

	// Registered clients.
	clients map[*Client]bool

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
	clients:    make(map[*Client]bool),
}

// getHub returns the hub
func getHub() *Hub {
	return &hub
}

// run starts the hub
func (h *Hub) run() {

	// Loop indefinitely
	for {

		//If any of these communication received (through these channels)
		select {

		// Register request from a client
		case client := <-h.register:

			// Add client to the registered clients map of hub
			h.clients[client] = true

		// Unregister request from a client
		case client := <-h.unregister:

			// If the current client exists in registered client map
			if _, ok := h.clients[client]; ok {

				// Delete the client from registered client map and close the channel
				delete(h.clients, client)
				close(client.send)
			}

		// Inbound message from a client
		case message := <-h.broadcast:

			// Range over all the currently registed clients (later, here will filter by message recipient as well)
			for client := range h.clients {
				select {

				// Forward the message to the client
				case client.send <- message:

				// If the client is not accepting the message
				default:

					// Delete the client and close the channel
					delete(h.clients, client)
					close(client.send)
				}
			}
		}
	}
}
