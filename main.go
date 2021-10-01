package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type Session struct {
	Nickname   string
	Connection *websocket.Conn
}

var chatRoom []Session

func connect(connection *websocket.Conn, nickname string, messageType int) {
	var newConnection Session = Session{Nickname: string(nickname), Connection: connection}
	chatRoom = append(chatRoom, newConnection)

	err := connection.WriteMessage(messageType, []byte("Connection approved"))

	if err != nil {
		log.Println("Error writing message:", err)
	}
}

// Pass struct as a pointer instead ?
func broadcastMessage(connection *websocket.Conn, message messageStruct, messageType int) {
	for _, session := range chatRoom {
		err := session.Connection.WriteMessage(messageType, []byte(message.Message))
		if err != nil {
			fmt.Println(err)
		}
	}
}

type messageStruct struct {
	Connection    bool
	Disconnection bool
	Nickname      string
	Message       string
}

func handleSession(w http.ResponseWriter, r *http.Request) {
	connection, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("Error intializing connection:", err)
		return
	}
	defer connection.Close()

	for {
		messageType, message, err := connection.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}

		var resp messageStruct
		json.Unmarshal(message, &resp)

		if resp.Connection {
			connect(connection, resp.Nickname, messageType)
		} else if resp.Disconnection {
		} else {
			broadcastMessage(connection, resp, messageType)
		}

		if err != nil {
			log.Println("Error writing message:", err)
			break
		}
	}
}

func main() {
	var port string
	flag.StringVar(&port, "port", "8080", "wrong port value")

	var addr string = "localhost:" + port
	http.HandleFunc("/run/session", handleSession)

	fmt.Println("Server running on port:", port)
	http.ListenAndServe(addr, nil)
}
