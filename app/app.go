package app

import (
	"fmt"
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"

	"websocket_test_1/controllers"
)

// StartApp starts the service
func StartApp() {

	// Start the hub in a new goroutine
	go controllers.GetHub().Run()

	// Start the server

	router := httprouter.New()

	router.HandlerFunc(http.MethodGet, "/chat/from/:userProfileId", http.HandlerFunc(controllers.HandleWebsocket)) //TODO: For now using unauthenticated route. Later on, will use auth middleware to authenticate and extract userIds

	fmt.Println("Starting web socket server at 8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}
