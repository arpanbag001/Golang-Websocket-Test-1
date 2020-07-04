package main

type message struct {
	data []byte
	room string
}

type subscription struct {
	conn *connection
	room string
}

// Hub maintains the set of active connections and broadcasts messages to the
// connections.
type Hub struct {
	// Registered connections.
	rooms map[string]map[*connection]bool

	// Inbound messages from the connections.
	broadcast chan message

	// Register requests from the connections.
	register chan subscription

	// Unregister requests from connections.
	unregister chan subscription
}

var hub = Hub{
	broadcast:  make(chan message),
	register:   make(chan subscription),
	unregister: make(chan subscription),
	rooms:      make(map[string]map[*connection]bool),
}

func getHub() *Hub{
	return &hub
}

func (hub *Hub) run() {

	//Loop indefinitely
	for {
		select {
		case s := <-hub.register:
			connections := hub.rooms[s.room]
			if connections == nil {
				connections = make(map[*connection]bool)
				hub.rooms[s.room] = connections
			}
			hub.rooms[s.room][s.conn] = true
		case s := <-hub.unregister:
			connections := hub.rooms[s.room]
			if connections != nil {
				if _, ok := connections[s.conn]; ok {
					delete(connections, s.conn)
					close(s.conn.send)
					if len(connections) == 0 {
						delete(hub.rooms, s.room)
					}
				}
			}
		case m := <-hub.broadcast:
			connections := hub.rooms[m.room]
			for c := range connections {
				select {
				case c.send <- m.data:
				default:
					close(c.send)
					delete(connections, c)
					if len(connections) == 0 {
						delete(hub.rooms, m.room)
					}
				}
			}
		}
	}
}
