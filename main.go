package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

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

type ConnectionResponse struct {
	IsAllowed bool
	Message   string
}

type MessageStruct struct {
	Connection    bool
	Disconnection bool
	Nickname      string
	Message       string
}

var chatRoom []Session

func isConnectionAllowed(nickname string) (bool, string) {
	if len(chatRoom) >= 10 {
		return false, "Connection denied: the server is full, try again later."
	}

	for _, sess := range chatRoom {
		if sess.Nickname == nickname {
			return false, "Connection denied: this nickname is already taken."
		}
	}
	return true, "Connection successful."
}

func disconnect(connection *websocket.Conn, nickname string, messageType int) {
	for i, session := range chatRoom {
		if session.Nickname == nickname {
			fmt.Println("INDEX", i)
			if len(chatRoom) == 1 {
				chatRoom = nil
			} else if i == 0 {
				chatRoom[i] = chatRoom[len(chatRoom)-1]
				chatRoom = chatRoom[:len(chatRoom)-1]
			} else {
				chatRoom[i] = chatRoom[0]
				chatRoom = chatRoom[:len(chatRoom)-1]
			}
			broadcastMessage(connection, nickname, "*"+nickname+" left the chat*", messageType)
		}
	}
}

func connect(connection *websocket.Conn, nickname string, messageType int) {
	isAllowed, errMessage := isConnectionAllowed(nickname)
	var response ConnectionResponse = ConnectionResponse{IsAllowed: isAllowed, Message: errMessage}

	marshaledPayload, error := json.Marshal(response)
	if error != nil {
		fmt.Println("error marshalling", error)
	}
	err := connection.WriteMessage(websocket.TextMessage, []byte(string(marshaledPayload)))
	if err != nil {
		log.Println("Error writing message.", err)
		return
	}

	if isAllowed {
		// Pushing the connection in the room
		var newConnection Session = Session{Nickname: string(nickname), Connection: connection}
		chatRoom = append(chatRoom, newConnection)
		broadcastMessage(connection, nickname, "*Has joined the chat*", messageType)

	} else {
		fmt.Printf("Connection not allowed")
	}
}

// Pass struct as a pointer instead ?
func broadcastMessage(connection *websocket.Conn, nickname string, message string, messageType int) {
	for _, session := range chatRoom {
		if session.Nickname != nickname {
			err := session.Connection.WriteMessage(messageType, []byte(nickname+": "+message))
			if err != nil {
				fmt.Println(err)
			}
		}
	}
}

func disconnectRoom() {
	for _, session := range chatRoom {
		err := session.Connection.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "Closing client: the server has shut down."))
		if err != nil {
			fmt.Println("Error on disconnecting user:", err)
		}
	}
	os.Exit(0)
}

func handleSession(w http.ResponseWriter, r *http.Request) {
	connection, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("Error intializing connection:", err)
		return
	}
	defer connection.Close()

	// CATCHING ^C
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for sig := range c {
			if sig == syscall.SIGINT {
				disconnectRoom()
			}
		}
	}()

	for {
		messageType, message, err := connection.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}

		var resp MessageStruct
		json.Unmarshal(message, &resp)

		if resp.Connection {
			connect(connection, resp.Nickname, messageType)
		} else if resp.Disconnection {
			disconnect(connection, resp.Nickname, messageType)
		} else {
			broadcastMessage(connection, resp.Nickname, resp.Message, messageType)
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
