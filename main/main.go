package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func main() {

	// Start the hub in a new goroutine
	go getHub().run()

	// Start the server

	router := httprouter.New()

	router.HandlerFunc(http.MethodGet, "/chat", http.HandlerFunc(handleWebsocket))	//TODO: For now using unauthenticated route. Later on, will use auth middleware to authenticate and extract userIds

	fmt.Println("Starting web socket server at 8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}
