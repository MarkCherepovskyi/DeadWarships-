//server.go
package main

import (
	"game/game"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

type ConnectUser struct {
	Websocket *websocket.Conn
	ClienID   string
}

var U = game.User{}
var users = make(map[ConnectUser]int)
var upgrader = websocket.Upgrader{} // use default options

func newConnectUser(conn *websocket.Conn, clientID string) *ConnectUser {
	game.AddPlayer(clientID)

	conn.WriteMessage(websocket.TextMessage, []byte(clientID))

	return &ConnectUser{
		Websocket: conn,
		ClienID:   clientID,
	}
}

func socketHandler(w http.ResponseWriter, r *http.Request) {
	// Upgrade our raw HTTP connection to a websocket based one
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("Error during connection upgradation:", err)
		return
	}
	log.Println("Client connected: ", conn.RemoteAddr().String())

	var socketClient *ConnectUser = newConnectUser(conn, conn.RemoteAddr().String())
	users[*socketClient] = 0
	log.Println("Number client conected: ", len(users))

	defer conn.Close()
	// The event loopы
	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			log.Println("Error during message reading:", err)
			break
		}
		log.Printf("Received:%s  %s", socketClient.ClienID, message)
		err = conn.WriteMessage(messageType, message)
		if err != nil {
			log.Println("Error during message writing:", err)
			break
		}

		//обход всех пользователей и передача им сообщений
		for client := range users {
			if client.ClienID != socketClient.ClienID {
				if err := client.Websocket.WriteMessage(messageType, message); err != nil {
					log.Println("Cloud not send message to ", client.ClienID, err.Error())
				}
			}
		}
	}
}

func main() {
	http.HandleFunc("/socket", socketHandler)
	log.Println("server is ready")
	log.Fatal(http.ListenAndServe("localhost:9999", nil))

}
