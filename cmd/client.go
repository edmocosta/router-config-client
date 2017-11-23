package main

import (
	"fmt"
	"golang.org/x/net/websocket"
	"log"
	"net/http"
	"time"
	"github.com/edmocosta/router-config-client/app/ws"
)

func main() {
	fmt.Println("Starting configuration client...")
	http.DefaultClient.Timeout = 5 * time.Second

	mux := http.NewServeMux()
	mux.Handle("/ws", websocket.Handler(ws.ConfigHandler))

	fmt.Println("\n")
	fmt.Println("Keep this client opened and back to the web page to congigure your router")
	log.Fatal(http.ListenAndServe(":1510", mux))
}
