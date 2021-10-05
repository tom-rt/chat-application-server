# Chat application server

Server of a chat application.

## Description

The server is meant to work with [clients](https://github.com/tom-rt/chat-application-client).
When launched, the server will wait for clients to connect and broadcast received messages.

## Getting Started

### Dependencies

* Go version 1.17
* [Gorilla websocket](github.com/gorilla/websocket) as the only external dependency.

### Installing

* To install the program:
```
git clone git@github.com:tom-rt/chat-application-server.git
cd chat-application-server
go get .
```

### Executing program

* There is one non mandatory arguments: port, which represent the port on whiche the server will listen. Its default value is 8080.
* The server can be started by running main.go file, example:
```
go run main.go
```

* Or by building and executing a binary, example:
```
go build
./server
```

## Author

https://github.com/tom-rt
