package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

func handleWebsocket(w http.ResponseWriter, r *http.Request) {

	//Check origin of the request. For now, allowing every request by returning true without checking
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }

	//Upgrade the request to websocket
	conn, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		log.Println(err)
		return
	}

	fmt.Println("Client successfully connected!")

	//Defer closing the connection
	defer func() {
		conn.Close()
		fmt.Println("Client disconnected!")
	}()

	//Read and write messages
	handleMessages(conn)
}

func handleMessages(conn *websocket.Conn) {

	//Loop indefinitely
	for {

		//Read the incoming message
		messageType, p, err := conn.ReadMessage()

		if err != nil {
			log.Println(err)
			return
		}

		//Print the received message
		log.Println(string(p))

		//Send the same message
		err = conn.WriteMessage(messageType, p)

		if err != nil {
			log.Println(err)
			return
		}
	}
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}
