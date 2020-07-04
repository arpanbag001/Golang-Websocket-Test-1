package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func main() {

	router := httprouter.New()

	router.HandlerFunc(http.MethodGet,"/chat/:roomId", http.HandlerFunc(handleWebsocket))
	fmt.Println("Starting web socket server at 8080")
	go getHub().run()
	log.Fatal(http.ListenAndServe(":8080", router))
}
