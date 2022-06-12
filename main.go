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
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

const (
	screenWidth  = 800
	screenHeight = 400
)

var (
	Rules = false
	win   bool
	lose  bool

	MyID          string
	usersInServer game.Users
	msg           = `When the window opens - you need to place the ships. To locate, right-click on the desired cell. 
	Also keep in mind that ships must be placed in a certain order: 4x, 3x, 3x, 2x, 2x, and 3 single-deck.
	 After that, the program will start sending requests to the server and wait for a response.
	  When the opponent also deploys all the ships - the game will begin.
	Each player takes turns shooting at the opponent's ships. To do this,
	 use the left mouse button, which is pressed on the desired cell.
	  If the player is simply wounded, the injured deck is blue, if the ship sank (Like Moscow) - it is painted black.`
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
	} else if ebiten.IsMouseButtonPressed(ebiten.MouseButtonMiddle) {
		Rules = !Rules

	}
	return nil

}

func (g *Game) Draw(screen *ebiten.Image) {

	for x := 0; x < 10; x++ {
		for y := 0; y < 10; y++ {
			game.DrawAllPlace(x, y, screen, usersInServer)

		}

	}
	if Rules {
		ebitenutil.DebugPrintAt(screen, msg, 20, 270)
	}

	if win {
		ebitenutil.DebugPrintAt(screen, " U win", 300, 20)
	} else if lose {
		ebitenutil.DebugPrintAt(screen, " U lose", 300, 20)
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

func checkEnemyDead() {
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
		lose = true

		return
	}
}

func checkMyDead() {
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

			bufferInArray := true
			for _, data := range game.UsersInServer[MyID].DeadWarships {
				if data == i {
					bufferInArray = false
					break
				}
			}
			if bufferInArray {
				game.UsersInServer[MyID].DeadWarships = append(game.UsersInServer[MyID].DeadWarships, i)
			}

		}

		log.Println("Num of dead warship", len(game.UsersInServer[MyID].DeadWarships))
	}
	////////////
	if len(game.UsersInServer[MyID].DeadWarships) == 8 {

		log.Println("U win")
		win = true
		return

	}
}

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
			go checkEnemyDead()
			if game.UsersInServer[MyID].NumberOfMyWarship >= 8 && game.UsersInServer[MyID].UserID != bufferUser.UserID {

				game.UsersInServer[MyID].CanMove = !bufferUser.CanMove

				game.UsersInServer[MyID].EnemyMoveX = bufferUser.LastMoveX
				game.UsersInServer[MyID].EnemyMoveY = bufferUser.LastMoveY

				game.UsersInServer[MyID].EnemyWarships = bufferUser.MyWarships

				bufferReadX = bufferUser.LastMoveX
				bufferReadY = bufferUser.LastMoveY

				game.EnemyMove(usersInServer)

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
				////////////////
				go checkMyDead()
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
