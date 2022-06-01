package main

import (
	"encoding/json"
	_ "image/png"
	"log"
	"os"
	"os/signal"
	"time"

	"game/game"

	"github.com/gorilla/websocket"
	ebiten "github.com/hajimehoshi/ebiten/v2"
)

const (
	screenWidth  = 800
	screenHeight = 400
)

var (
	MyID          string
	usersInServer game.Users
)

type Game struct{}

func (g *Game) Update() error {

	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {

		game.Move(usersInServer)

	} else if ebiten.IsMouseButtonPressed(ebiten.MouseButtonRight) {
		game.PlacingMyWarships(usersInServer)
	}
	return nil

}

func (g *Game) Draw(screen *ebiten.Image) {

	for x := 0; x < 10; x++ {
		for y := 0; y < 10; y++ {
			game.DrawAllPlace(x, y, screen, usersInServer)

		}

	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func initialiseGame() {
	//MyID = game.MyID

	game.AddPlayer(MyID)

	for {
		usersInServer = game.UsersInServer
		if len(usersInServer) > 0 {

			log.Println(usersInServer[MyID].UserID)
			log.Println(len(usersInServer))

			////////////initialise plase

			game.InitialPlace(usersInServer)
			game.InitialMyPlace(usersInServer)

			////////initialise warships

			//game.InitialEnemyWarships(	usersInServer)

			return
		}

		time.Sleep(1 * time.Second)
	}
	///////////

}

/////
var Ready bool
var done chan interface{}
var interrupt chan os.Signal
var bufferWriteX = 10000
var bufferWriteY = 10000
var bufferReadY int
var bufferReadX int

func receiveHandler(conn *websocket.Conn) {
	defer close(done)
	i := 0
	for {
		if MyID == "" && i == 0 {
			_, id, err := conn.ReadMessage()
			if err != nil {
				log.Println("All are bad !!!!!!!!!1", err)
			}
			MyID = string(id)
			log.Println("MyID", MyID)
			i++
		} else {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				log.Println("Error in receive:", err)
				return
			}

			bufferUser := game.User{}

			json.Unmarshal(msg, &bufferUser)
			//&& bufferReadX != bufferUser.LastMoveX && bufferReadY != bufferUser.LastMoveY
			if game.UsersInServer[MyID].NumberOfMyWarship >= 10 && game.UsersInServer[MyID].UserID != bufferUser.UserID && usersInServer[MyID].CanMove == false {
				game.UsersInServer[MyID].CanMove = true
				//game.UsersInServer[MyID].CanMove = !bufferUser.CanMove

				game.UsersInServer[MyID].EnemyMoveX = bufferUser.LastMoveX
				game.UsersInServer[MyID].EnemyMoveY = bufferUser.LastMoveY

				game.UsersInServer[MyID].EnemyWarships = bufferUser.MyWarships
				//log.Printf("Received: %s\n", msg)
				bufferReadX = bufferUser.LastMoveX
				bufferReadY = bufferUser.LastMoveY

				game.EnemyMove(usersInServer)
				log.Println("MYID game", game.UsersInServer[MyID].UserID)
				log.Println("my id", MyID)
				log.Println("EnemyID", bufferUser.UserID)
				log.Println("enemy warships ", bufferUser.EnemyWarships)
				log.Println("my wrships ", bufferUser.MyWarships)
				log.Println("user can move ", game.UsersInServer[MyID].CanMove)
			}

			//og.Println("num warship ", bufferUser.NumberOfMyWarship)
			//log.Println("num warship2 ", game.UsersInServer[MyID].NumberOfMyWarship)

			//log.Println("Received: ", msg)
		}

	}
}

func main() {
	done = make(chan interface{})    // Channel to indicate that the receiverHandler is done
	interrupt = make(chan os.Signal) // Channel to listen for interrupt signal to terminate gracefully

	signal.Notify(interrupt, os.Interrupt) // Notify the interrupt channel for SIGINT

	socketUrl := "ws://localhost:9999" + "/socket"
	conn, _, err := websocket.DefaultDialer.Dial(socketUrl, nil)
	if err != nil {
		log.Fatal("Error connecting to Websocket Server:", err)
	}
	defer conn.Close()
	go receiveHandler(conn)

	// Our main loop for the client
	// We send our relevant packets here
	go func(conn *websocket.Conn) {
		for {
			select {
			case <-time.After(time.Duration(1) * time.Millisecond * 1000):
				//var msg string
				//fmt.Fscan(os.Stdin, &msg) //write msg
				//log.Println("nunber of my warship ", game.UsersInServer[MyID].NumberOfMyWarship)
				if game.UsersInServer[MyID].NumberOfMyWarship >= 10 && game.UsersInServer[MyID].LastMoveX != bufferWriteX && game.UsersInServer[MyID].LastMoveY != bufferWriteY {

					bufferOfUserForSend, _ := json.Marshal(game.UsersInServer[MyID])
					err := conn.WriteMessage(websocket.TextMessage, []byte(bufferOfUserForSend))
					//err := conn.WriteMessage(websocket.TextMessage, []byte("sdfsd"))
					if err != nil {
						log.Println("Error during writing to websocket:", err)
						return
					}
					bufferWriteX = game.UsersInServer[MyID].LastMoveX
					bufferWriteY = game.UsersInServer[MyID].LastMoveY
					continue
				}

				/*err := conn.WriteMessage(websocket.TextMessage, []byte("bufferOfUserForSend"))
				//err := conn.WriteMessage(websocket.TextMessage, []byte("sdfsd"))
				if err != nil {
					log.Println("Error during writing to websocket:", err)
					return
				}*/

			case <-interrupt:
				// We received a SIGINT (Ctrl + C). Terminate gracefully…
				log.Println("Received SIGINT interrupt signal. Closing all pending connections")

				// Close our websocket connection
				err := conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
				if err != nil {
					log.Println("Error during closing websocket:", err)
					return
				}

				select {
				case <-done:
					log.Println("Receiver Channel Closed! Exiting….")
				case <-time.After(time.Duration(1) * time.Second):
					log.Println("Timeout in closing receiving channel. Exiting….")
				}
				return
			}
		}
	}(conn)

	ebiten.SetWindowSize(screenWidth, screenHeight)

	ebiten.SetWindowTitle("My Game")
	log.Println("My ID ", MyID)
	go initialiseGame()
	if err := ebiten.RunGame(&Game{}); err != nil {
		log.Fatal(err)
	}
}
