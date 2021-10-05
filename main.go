package main

import (
	"chat-application/server/handlers/logging"
	"chat-application/server/handlers/session"
	"flag"
	"fmt"
	"log"
	"net/http"
)

func main() {
	var port string
	flag.StringVar(&port, "port", "8080", "wrong port value")
	var addr string = "localhost:" + port

	err := logging.InitLog()
	if err != nil {
		log.Println("Cannot open or create log file:", err)
		return
	}

	http.HandleFunc("/run/session", session.HandleSession)

	fmt.Println("Server running on port:", port)
	http.ListenAndServe(addr, nil)
}
