package main

import (
	"fmt"
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
	usersInServer *game.Users
)

type Game struct{}

func (g *Game) Update() error {

	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {

		game.Move(*usersInServer)

	} else if ebiten.IsMouseButtonPressed(ebiten.MouseButtonRight) {
		game.PlacingMyWarships(*usersInServer)
	}
	return nil

}

func (g *Game) Draw(screen *ebiten.Image) {

	for x := 0; x < 10; x++ {
		for y := 0; y < 10; y++ {
			game.DrawAllPlace(x, y, screen, *usersInServer)

		}

	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func initialiseGame() {
	MyID = game.MyID

	game.AddPlayer(MyID)

	for {
		usersInServer = &game.UsersInSerwer
		if len(*usersInServer) > 0 {
			////////////initialise plase

			game.InitialPlace(*usersInServer)
			game.InitialMyPlace(*usersInServer)

			////////initialise warships

			game.InitialEnemyWarships(*usersInServer)

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

func receiveHandler(conn *websocket.Conn) {
	defer close(done)
	for {
		if MyID == "" {
			_, id, err := conn.ReadMessage()
			if err != nil {
				log.Println("All are bad !!!!!!!!!1", err)
			}
			MyID = string(id)
			log.Println("MyID", MyID)
		}

		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Println("Error in receive:", err)
			return
		}
		log.Printf("Received: %s\n", msg)
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

	///////////////

	//////////////////

	// Our main loop for the client
	// We send our relevant packets here
	go func(conn *websocket.Conn) {
		for {
			select {
			case <-time.After(time.Duration(1) * time.Millisecond * 1000):
				var msg string
				fmt.Fscan(os.Stdin, &msg) //write msg

				err := conn.WriteMessage(websocket.TextMessage, []byte(msg))
				if err != nil {
					log.Println("Error during writing to websocket:", err)
					return
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
