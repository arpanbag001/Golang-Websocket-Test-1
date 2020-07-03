package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/chat", handleWebsocket)
	fmt.Println("Starting web socket server at 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
