package main

import (
	"encoding/json"
	"image/color"
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
		for i := range game.UsersInServer[MyID].MyWarships {
			bufferSize := 0
			for _, data := range game.UsersInServer[MyID].MyWarships[i] {
				if game.UsersInServer[MyID].ArrayEnemyPlace[(data[0]*10)+data[1]].WasShot {
					bufferSize++
				}
			}
			if bufferSize == len(game.UsersInServer[MyID].MyWarships[i]) {
				log.Println("I kill 1 warships")

				for _, data := range game.UsersInServer[MyID].MyWarships[i] {
					game.UsersInServer[MyID].ArrayEnemyPlace[(data[0]*10)+data[1]].ShotWarship(color.RGBA{0, 50, 50, 255})
				}
			}

		}

		for i := range game.UsersInServer[MyID].EnemyWarships {
			bufferSize := 0
			for _, data := range game.UsersInServer[MyID].EnemyWarships[i] {
				if game.UsersInServer[MyID].ArrayMyPlace[(data[0]*10)+data[1]].WasShot {
					bufferSize++
				}
			}
			if bufferSize == len(game.UsersInServer[MyID].EnemyWarships[i]) {
				log.Println("I kill 1 warships")
				for _, data := range game.UsersInServer[MyID].EnemyWarships[i] {
					game.UsersInServer[MyID].ArrayMyPlace[(data[0]*10)+data[1]].ShotWarship(color.RGBA{0, 50, 50, 255})
				}
			}

		}
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
				log.Println("I don't know who are you", err)
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

			if game.UsersInServer[MyID].NumberOfMyWarship >= 8 && game.UsersInServer[MyID].UserID != bufferUser.UserID {

				bufferForAll := 0
				for i := range game.UsersInServer[MyID].MyWarships {
					bufferSize := 0
					for _, data := range game.UsersInServer[MyID].MyWarships[i] {
						if game.UsersInServer[MyID].ArrayMyPlace[(data[0]*10)+data[1]].WasShot {
							bufferSize++
						}
					}
					if bufferSize == len(game.UsersInServer[MyID].MyWarships[i]) {
						for _, data := range game.UsersInServer[MyID].MyWarships[i] {
							game.UsersInServer[MyID].ArrayMyPlace[(data[0]*10)+data[1]].Kill()
							bufferForAll++
						}

					}
				}

				//if len(bufferUser.DeadWarships) == 8 {
				if bufferForAll == 8 {
					log.Println("U lose")
					return
				}

				game.UsersInServer[MyID].CanMove = !bufferUser.CanMove

				game.UsersInServer[MyID].EnemyMoveX = bufferUser.LastMoveX
				game.UsersInServer[MyID].EnemyMoveY = bufferUser.LastMoveY

				game.UsersInServer[MyID].EnemyWarships = bufferUser.MyWarships

				bufferReadX = bufferUser.LastMoveX
				bufferReadY = bufferUser.LastMoveY

				game.EnemyMove(usersInServer)

				/*//log.Println("MYID game", game.UsersInServer[MyID].UserID)
				//log.Println("my id", MyID)
				//log.Println("EnemyID", bufferUser.UserID)
				log.Println("enemy warships ", bufferUser.EnemyWarships)
				log.Println("my wrships ", bufferUser.MyWarships)*/
				log.Println("user can move ", game.UsersInServer[MyID].CanMove)
			}

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
			case <-time.After(time.Duration(1) * time.Millisecond * 100):
				bufferForAll := 0
				for i := range game.UsersInServer[MyID].EnemyWarships {
					bufferSize := 0
					for _, data := range game.UsersInServer[MyID].EnemyWarships[i] {
						if game.UsersInServer[MyID].ArrayEnemyPlace[(data[0]*10)+data[1]].WasShot {
							bufferSize++
						}
					}
					if bufferSize == len(game.UsersInServer[MyID].EnemyWarships[i]) {
						for _, data := range game.UsersInServer[MyID].EnemyWarships[i] {
							game.UsersInServer[MyID].ArrayEnemyPlace[(data[0]*10)+data[1]].Kill()
						}

					}
				}
				if bufferForAll == 8 {

					log.Println("U win")
					return

				}
				if game.UsersInServer[MyID].NumberOfMyWarship >= 8 {

					bufferOfUserForSend, _ := json.Marshal(game.UsersInServer[MyID])
					err := conn.WriteMessage(websocket.TextMessage, []byte(bufferOfUserForSend))

					if err != nil {
						log.Println("Error during writing to websocket:", err)
						return
					}

					bufferWriteX = game.UsersInServer[MyID].LastMoveX
					bufferWriteY = game.UsersInServer[MyID].LastMoveY
					continue
				}

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
