package game

import (
	"fmt"
	"log"
	"math/rand"

	"image/color"

	ebiten "github.com/hajimehoshi/ebiten/v2"
)

type User struct {
	UserID                       string     `json:"id"`
	CanMove                      bool       `json:"canMove"`
	EnemyWarships                [10][2]int `json:"EnemyWarships"`
	MyWarships                   [10][2]int `json:"MyWarships"`
	arrayEnemyPlace              [100]Place //`json:"enemyPlace"`
	arrayMyPlace                 [100]Place //`json:"myPlace"`
	NumberOfMyWarship            int
	numberOfEnemyWarships        int
	updateBufferX, updateBufferY int
	LastMoveX                    int `json:"lastMoveX"`
	LastMoveY                    int `json:"lastMoveY"`
	EnemyMoveX, EnemyMoveY       int
}

type Users map[string]*User

var (
	UsersInServer = make(Users)
	MyID          = ""
)

type Place struct {
	localx, localy int
	size           int
	placeX, placeY int
	value          bool
	colorPlace     color.Color
}

func (p *Place) ShotWarship() error {
	p.colorPlace = color.RGBA{0, 0, 255, 255}
	return nil
}

func (p *Place) CreateWarships(bufferX, bufferY, num int, users Users) error {

	users[MyID].MyWarships[num][0] = bufferX
	users[MyID].MyWarships[num][1] = bufferY
	p.colorPlace = color.RGBA{0, 0, 255, 255}
	users[MyID].NumberOfMyWarship++

	return nil
}

func (p *Place) UpdatePlace() error {
	if !p.value {
		p.colorPlace = color.RGBA{250, 250, 250, 250}
		p.value = false

	}
	return nil
}

func (p *Place) DrawPlace(screen *ebiten.Image) error {
	if p.colorPlace == nil {
		p.colorPlace = color.RGBA{255, 0, 255, 255}
	}

	for i := p.placeX; i < p.placeX+p.size; i++ {
		for j := p.placeY; j < p.placeY+p.size; j++ {
			screen.Set(i, j, p.colorPlace)
		}
	}
	return nil
}

/**/

func Move(users Users) {
	if users[MyID].CanMove {
		mx, my := ebiten.CursorPosition()
		if mx >= 10 && mx <= 260 && my <= 260 && my >= 10 {
			bufferX := int((mx - 10) / 25)
			bufferY := int((my - 10) / 25)

			fmt.Println("You press button")

			for i := 0; i < 10; i++ {
				if users[MyID].arrayEnemyPlace[(bufferX*10)+bufferY].localx == users[MyID].EnemyWarships[i][0] && users[MyID].arrayEnemyPlace[(bufferX*10)+bufferY].localy == users[MyID].EnemyWarships[i][1] {
					users[MyID].arrayEnemyPlace[(bufferX*10)+bufferY].ShotWarship()
					fmt.Println("shot")
					return
				}

			}

			users[MyID].arrayEnemyPlace[(bufferX*10)+bufferY].UpdatePlace()

			users[MyID].LastMoveX = bufferX
			users[MyID].LastMoveY = bufferY

			users[MyID].CanMove = false
		}
	}

}

func EnemyMove(users Users) {
	//	if users[MyID].CanMove {
	enemyX := users[MyID].EnemyMoveX
	enemyY := users[MyID].EnemyMoveY
	for i := 0; i < 10; i++ {
		if users[MyID].arrayMyPlace[(enemyX*10)+enemyY].localx == users[MyID].MyWarships[i][0] && users[MyID].arrayMyPlace[(enemyX*10)+enemyY].localy == users[MyID].MyWarships[i][1] {
			users[MyID].arrayMyPlace[(enemyX*10)+enemyY].ShotWarship()
			fmt.Println("shot by me")
			return
		}

	}

	users[MyID].arrayMyPlace[(enemyX*10)+enemyY].UpdatePlace()

	//users[MyID].CanMove = !users[MyID].CanMove
	//}//

}

func PlacingMyWarships(users Users) {
	if users[MyID].NumberOfMyWarship < 10 {
		mx, my := ebiten.CursorPosition()

		if mx >= 400 && mx <= 650 && my <= 260 && my >= 10 {

			bufferX := int((mx - 400) / 25)
			bufferY := int((my - 10) / 25)
			fmt.Println("You press button1111")
			if bufferX != users[MyID].updateBufferX || bufferY != users[MyID].updateBufferY {
				users[MyID].arrayMyPlace[(bufferX*10)+bufferY].CreateWarships(bufferX, bufferY, users[MyID].NumberOfMyWarship, users)

			}
			users[MyID].updateBufferX = bufferX
			users[MyID].updateBufferY = bufferY

		}
	} /* else {
		users[MyID].CanMove = true

	}*/
}

func InitialPlace(users Users) [100]Place {
	placeX, placeY := 10, 10

	purpleCol := color.RGBA{255, 0, 255, 255} //настройка цвета для текста

	for x := 0; x < 10; x++ {
		for y := 0; y < 10; y++ {
			users[MyID].arrayEnemyPlace[(y*10)+x] = Place{x, y, 20, placeX, placeY, false, purpleCol}
			placeX += 25

		}
		placeY += 25
		placeX = 10
	}
	return users[MyID].arrayEnemyPlace
}

func InitialEnemyWarships(users Users) [10][2]int {
	for i := 0; i < users[MyID].numberOfEnemyWarships; i++ {
		for j := 0; j < 2; j++ {
			users[MyID].EnemyWarships[i][j] = rand.Intn(10)
			fmt.Println("war...... x: ", users[MyID].EnemyWarships[i][0], " war...... y: ", users[MyID].EnemyWarships[i][1])

		}
	}
	return users[MyID].EnemyWarships
}

func InitialMyPlace(users Users) [100]Place {
	purpleCol := color.RGBA{255, 0, 255, 255} //настройка цвета для текста
	placeX, placeY := 400, 10
	for x := 0; x < 10; x++ {
		for y := 0; y < 10; y++ {

			users[MyID].arrayMyPlace[(y*10)+x] = Place{x, y, 20, placeX, placeY, false, purpleCol}
			placeX += 25

		}
		placeY += 25
		placeX = 400
	}
	return users[MyID].arrayMyPlace
}

func DrawAllPlace(x, y int, screen *ebiten.Image, users Users) {
	users[MyID].arrayEnemyPlace[(x*10)+y].DrawPlace(screen)
	users[MyID].arrayMyPlace[(x*10)+y].DrawPlace(screen)
}

///////

func AddPlayer(userID string) *User {

	user := &User{
		UserID:                userID,
		CanMove:               false,
		numberOfEnemyWarships: 10,
		NumberOfMyWarship:     0,
	}

	MyID = userID
	log.Println("player add", MyID)

	UsersInServer[userID] = user

	return user
}
