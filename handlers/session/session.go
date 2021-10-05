package session

import (
	"chat-application/server/handlers/logging"
	"encoding/json"
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

type MessageStruct struct {
	Connection    bool
	Disconnection bool
	Nickname      string
	Message       string
}

type Session struct {
	Nickname   string
	Connection *websocket.Conn
}

type ConnectionResponse struct {
	IsAllowed bool
	Message   string
}

var logFile *os.File
var chatRoom []Session

func isConnectionAllowed(chatRoom []Session, nickname string) (bool, string) {
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
			logging.WriteLog(nickname + " has left the chat\n")
		}
	}
}

func connect(connection *websocket.Conn, nickname string, messageType int) {
	isAllowed, errMessage := isConnectionAllowed(chatRoom, nickname)
	var response ConnectionResponse = ConnectionResponse{IsAllowed: isAllowed, Message: errMessage}

	marshaledPayload, error := json.Marshal(response)
	if error != nil {
		log.Println("error marshalling:", error)
	}
	err := connection.WriteMessage(websocket.TextMessage, []byte(string(marshaledPayload)))
	if err != nil {
		log.Println("error writing message:", err)
		return
	}

	if isAllowed {
		var newConnection Session = Session{Nickname: string(nickname), Connection: connection}
		chatRoom = append(chatRoom, newConnection)
		broadcastMessage(connection, nickname, "*Has joined the chat*", messageType)
		logging.WriteLog(nickname + " has joined the chat\n")
	} else {
		fmt.Printf("Connection not allowed")
	}
}

func disconnectRoom() {
	for _, session := range chatRoom {
		err := session.Connection.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "Closing client: the server has shut down."))
		if err != nil {
			log.Println("error on disconnecting user:", err)
		}
	}
	logFile.Close()
	os.Exit(0)
}

func broadcastMessage(connection *websocket.Conn, nickname string, message string, messageType int) {
	for _, session := range chatRoom {
		if session.Nickname != nickname {
			err := session.Connection.WriteMessage(messageType, []byte(nickname+": "+message))
			if err != nil {
				log.Println("error writing message", err)
			}
		}
	}
}

func catchCtrlC() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	for sig := range c {
		if sig == syscall.SIGINT {
			disconnectRoom()
		}
	}
}

func HandleSession(w http.ResponseWriter, r *http.Request) {
	connection, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("error intializing connection:", err)
		return
	}
	defer connection.Close()

	go catchCtrlC()

	for {
		messageType, message, err := connection.ReadMessage()
		if err != nil {
			log.Println("error reading message:", err)
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
			log.Println("error writing message:", err)
			break
		}
	}
}
