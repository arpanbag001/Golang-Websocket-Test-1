package main

type message struct {
	data []byte
	roomID string
}

type subscription struct {
	conn *connection
	roomID string
}

// Hub maintains the set of active connections and broadcasts messages to the
// connections.
type Hub struct {
	// Registered connections.
	rooms map[string]map[*connection]bool

	// Inbound messages from any connection
	broadcast chan message

	// Register requests from any connection
	register chan subscription

	// Unregister requests from any connection
	unregister chan subscription
}

var hub = Hub{
	broadcast:  make(chan message),
	register:   make(chan subscription),
	unregister: make(chan subscription),
	rooms:      make(map[string]map[*connection]bool),
}

func getHub() *Hub {
	return &hub
}

func (hub *Hub) run() {

	//Loop indefinitely
	for {

		//If any of these communication received (through these channels)
		select {

		// Register requests from the connections.
		case s := <-hub.register:

			//Get all the connections of the room
			connections := hub.rooms[s.roomID]

			//If there is no connection in the room, indicating the room is empty
			if connections == nil {

				//Create the connection map
				connections = make(map[*connection]bool)
				hub.rooms[s.roomID] = connections
			}

			//Add the current connection to the room
			hub.rooms[s.roomID][s.conn] = true

		// Unregister requests from connections.
		case s := <-hub.unregister:

			//Get all the connections of the room
			connections := hub.rooms[s.roomID]

			if connections != nil {

				//If current connection exists in the room
				if _, ok := connections[s.conn]; ok {
					closeConnectionAndDeleteFromRoom(s.roomID, s.conn)
				}
			}

		// Inbound messages from any connection
		case msg := <-hub.broadcast:

			//Get all the connections of the room
			connections := hub.rooms[msg.roomID]

			//Iterate over all the connections of the room
			for conn := range connections {
				select {

				//Forward the message data to the connection
				case conn.send <- msg.data:

				//If the connection is not ready to receive the message
				default:

					//Delete the connection from the room and close the connection
					closeConnectionAndDeleteFromRoom(msg.roomID, conn)
				}
			}
		}
	}
}

//closeConnectionAndDeleteFromRoom closes the connection and deletes it from its room
func closeConnectionAndDeleteFromRoom(roomID string, connToClose *connection){

	//Get all the connections of the room
	connections := hub.rooms[roomID]

	//Delete the connection from the room and close the connection
	delete(connections, connToClose)
	close(connToClose.send)

	//If there is no connection in the room, indicating the room is empty
	if len(connections) == 0 {

		//Delete the room
		delete(hub.rooms, roomID)
	}
}