package game

import (
	"log"

	"image/color"

	ebiten "github.com/hajimehoshi/ebiten/v2"
)

type User struct {
	UserID                       string     `json:"id"`
	CanMove                      bool       `json:"canMove"`
	EnemyWarships                [][][]int  `json:"EnemyWarships"`
	MyWarships                   [][][]int  `json:"MyWarships"`
	ArrayEnemyPlace              [100]Place //`json:"enemyPlace"`
	ArrayMyPlace                 [100]Place //`json:"myPlace"`
	NumberOfMyWarship            int
	numberOfEnemyWarships        int
	updateBufferX, updateBufferY int
	LastMoveX                    int `json:"lastMoveX"`
	LastMoveY                    int `json:"lastMoveY"`
	EnemyMoveX, EnemyMoveY       int
	DeadWarships                 []int //how many I kill enemy warships
	DeadMyWarships               []int //how many enemy kill my warships
}

type Users map[string]*User

var (
	UsersInServer         = make(Users)
	MyID                  = ""
	bufferNumOfMyWarships int
	bufferSizeOfMyWarship int
)

type Place struct {
	localx, localy int
	size           int
	placeX, placeY int
	value          bool
	colorPlace     color.Color
	WasShot        bool
}

func (p *Place) ShotWarship(color color.Color) error {
	p.colorPlace = color
	p.WasShot = true
	return nil
}

func (p *Place) CreateWarships(bufferX, bufferY, num int, size int, users Users) error {

	users[MyID].MyWarships[num][size][0] = bufferX
	users[MyID].MyWarships[num][size][1] = bufferY
	log.Println("X", bufferX)
	log.Println("Y", bufferY)
	bufferSizeOfMyWarship++
	p.colorPlace = color.RGBA{0, 0, 255, 255}

	return nil
}

func (p *Place) UpdatePlace() error {
	if !p.value {
		p.colorPlace = color.RGBA{250, 250, 250, 250}
		p.value = false

	}
	return nil
}

func (p *Place) Kill() {
	p.colorPlace = color.RGBA{0, 50, 50, 1}
}

func (p *Place) DrawPlace(screen *ebiten.Image) error {
	if p.colorPlace == nil {
		p.colorPlace = color.RGBA{0, 0, 255, 255}
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

			for i := range users[MyID].MyWarships {

				for j := range users[MyID].MyWarships[i] {
					if users[MyID].ArrayEnemyPlace[(bufferX*10)+bufferY].localx == users[MyID].EnemyWarships[i][j][1] && users[MyID].ArrayEnemyPlace[(bufferX*10)+bufferY].localy == users[MyID].EnemyWarships[i][j][0] {
						users[MyID].ArrayEnemyPlace[(bufferX*10)+bufferY].ShotWarship(color.RGBA{0, 255, 255, 255})

						users[MyID].LastMoveX = bufferX
						users[MyID].LastMoveY = bufferY
						return
					}
				}
			}
			users[MyID].ArrayEnemyPlace[(bufferX*10)+bufferY].UpdatePlace()

			users[MyID].LastMoveX = bufferX
			users[MyID].LastMoveY = bufferY

			users[MyID].CanMove = false
		}
	}

}

func EnemyMove(users Users) {

	enemyX := users[MyID].EnemyMoveX
	enemyY := users[MyID].EnemyMoveY
	for i := range users[MyID].MyWarships {

		for j := range users[MyID].MyWarships[i] {
			if users[MyID].ArrayMyPlace[(enemyX*10)+enemyY].localx == users[MyID].MyWarships[i][j][1] && users[MyID].ArrayMyPlace[(enemyX*10)+enemyY].localy == users[MyID].MyWarships[i][j][0] {
				users[MyID].ArrayMyPlace[(enemyX*10)+enemyY].ShotWarship(color.RGBA{0, 255, 255, 255})
				return
			}
		}

	}

	users[MyID].ArrayMyPlace[(enemyX*10)+enemyY].UpdatePlace()

}

func PlacingMyWarships(users Users) {
	if users[MyID].NumberOfMyWarship < 8 {

		mx, my := ebiten.CursorPosition()
		if mx >= 400 && mx <= 650 && my <= 260 && my >= 10 {

			bufferX := int((mx - 400) / 25)
			bufferY := int((my - 10) / 25)

			if bufferX != users[MyID].updateBufferX || bufferY != users[MyID].updateBufferY {
				users[MyID].ArrayMyPlace[(bufferX*10)+bufferY].CreateWarships(bufferX, bufferY, bufferNumOfMyWarships, bufferSizeOfMyWarship, users)

			}
			if bufferSizeOfMyWarship == len(users[MyID].MyWarships[bufferNumOfMyWarships]) {
				bufferSizeOfMyWarship = 0
				bufferNumOfMyWarships++
				users[MyID].NumberOfMyWarship++
				log.Println("I add 1 warships")
			}
			users[MyID].updateBufferX = bufferX
			users[MyID].updateBufferY = bufferY

		}

	}
}

func InitialPlace(users Users) [100]Place {
	placeX, placeY := 10, 10

	purpleCol := color.RGBA{255, 0, 255, 255} //настройка цвета для текста

	for x := 0; x < 10; x++ {
		for y := 0; y < 10; y++ {
			users[MyID].ArrayEnemyPlace[(y*10)+x] = Place{x, y, 20, placeX, placeY, false, purpleCol, false}
			placeX += 25

		}
		placeY += 25
		placeX = 10
	}
	return users[MyID].ArrayEnemyPlace
}

/*
func InitialEnemyWarships(users Users) [10][2]int {
	for i := 0; i < users[MyID].numberOfEnemyWarships; i++ {
		for j := 0; j < 2; j++ {
			users[MyID].EnemyWarships[i][j] = rand.Intn(10)
			fmt.Println("war...... x: ", users[MyID].EnemyWarships[i][0], " war...... y: ", users[MyID].EnemyWarships[i][1])

		}
	}
	return users[MyID].EnemyWarships
}
*/

func InitialMyPlace(users Users) [100]Place {
	purpleCol := color.RGBA{255, 0, 255, 255} //настройка цвета для текста
	placeX, placeY := 400, 10
	for x := 0; x < 10; x++ {
		for y := 0; y < 10; y++ {
			users[MyID].ArrayMyPlace[(y*10)+x] = Place{x, y, 20, placeX, placeY, false, purpleCol, false}
			placeX += 25
		}
		placeY += 25
		placeX = 400
	}
	return users[MyID].ArrayMyPlace
}

func DrawAllPlace(x, y int, screen *ebiten.Image, users Users) {
	users[MyID].ArrayEnemyPlace[(x*10)+y].DrawPlace(screen)
	users[MyID].ArrayMyPlace[(x*10)+y].DrawPlace(screen)
}

///////

func AddPlayer(userID string) *User {

	user := &User{
		UserID:                userID,
		CanMove:               false,
		numberOfEnemyWarships: 8,
		NumberOfMyWarship:     0,
		DeadWarships:          make([]int, 0),
		DeadMyWarships:        make([]int, 0),
		MyWarships:            make([][][]int, 8),
	}

	user.MyWarships[0] = make([][]int, 4)
	user.MyWarships[1] = make([][]int, 3)
	user.MyWarships[2] = make([][]int, 3)
	user.MyWarships[3] = make([][]int, 2)
	user.MyWarships[4] = make([][]int, 2)
	user.MyWarships[5] = make([][]int, 1)
	user.MyWarships[6] = make([][]int, 1)
	user.MyWarships[7] = make([][]int, 1)

	for i := range user.MyWarships {
		for j := range user.MyWarships[i] {
			user.MyWarships[i][j] = make([]int, 2)
		}
	}

	MyID = userID
	log.Println("player add", MyID)

	UsersInServer[userID] = user

	return user
}
