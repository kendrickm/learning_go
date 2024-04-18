package game

import (
	"bufio"
	"fmt"
	"math"
	"os"
)

type Game struct {
	LevelChans []chan *Level
	InputChan  chan *Input

	Level *Level
}

func NewGame(numWindows int, path string) *Game {
	levelChans := make([]chan *Level, numWindows)
	for i := range levelChans {
		levelChans[i] = make(chan *Level)
	}

	inputChan := make(chan *Input)

	return &Game{levelChans, inputChan, loadLevelFromFile(path)}

}

type InputType int

const (
	None InputType = iota
	Up
	Down
	Left
	Right
	QuitGame
	CloseWindow
	Search // Temp
)

type Input struct {
	Typ          InputType
	LevelChannel chan *Level
}

type Tile rune

const (
	StoneWall  Tile = '#'
	DirtFloor  Tile = '.'
	ClosedDoor Tile = '|'
	OpenDoor   Tile = '/'
	Blank      Tile = 0
	Pending    Tile = -1
)

type Pos struct {
	X, Y int
}

type Entity struct {
	Pos
}

type Player struct {
	Entity
}

type Level struct {
	Map    [][]Tile
	Player Player
	Debug  map[Pos]bool
}

func loadLevelFromFile(filename string) *Level {
	file, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	levelLines := make([]string, 0)
	longestRow := 0
	index := 0
	for scanner.Scan() {
		// fmt.Println(scanner.Text())
		levelLines = append(levelLines, scanner.Text())
		if len(levelLines[index]) > longestRow {
			longestRow = len(levelLines[index])
		}
		index++
	}

	level := &Level{}
	level.Map = make([][]Tile, len(levelLines))
	for i := range level.Map {
		level.Map[i] = make([]Tile, longestRow)
	}

	for y := 0; y < len(level.Map); y++ {
		line := levelLines[y]

		for x, c := range line {
			var t Tile

			switch c {
			case ' ', '\n', '\t', '\r':
				t = Blank
			case '#':
				t = StoneWall
			case '|':
				t = ClosedDoor
			case '/':
				t = OpenDoor
			case '.':
				t = DirtFloor
			case 'P':
				level.Player.X = x
				level.Player.Y = y
				t = Pending
			default:
				panic("Invalid character in map")
			}
			level.Map[y][x] = t

		}
	}

	for y, row := range level.Map {
		for x, tile := range row {
			if tile == Pending {
			SearchLoop:
				for searchX := x - 1; searchX <= x+1; searchX++ {
					for searchY := y - 1; searchY <= y+1; searchY++ {
						searchTile := level.Map[searchY][searchX]
						switch searchTile {
						case DirtFloor:
							level.Map[y][x] = DirtFloor
							break SearchLoop
						}

					}
				}
			}
		}
	}
	return level
}
func canWalk(level *Level, pos Pos) bool {
	t := level.Map[pos.Y][pos.X]

	switch t {
	case StoneWall, ClosedDoor, Blank:
		return false
	default:
		return true
	}
}

func checkDoor(level *Level, pos Pos) {
	t := level.Map[pos.Y][pos.X]
	if t == ClosedDoor {
		level.Map[pos.Y][pos.X] = OpenDoor
	}
}

func (game *Game) handleInput(input *Input) {
	level := game.Level
	p := level.Player
	switch input.Typ {
	case Search:
		//game.bfs(, p.Pos)
		game.astar(p.Pos, Pos{3, 2})
	case Up:
		if canWalk(level, Pos{p.X, p.Y - 1}) {
			level.Player.Y--
		} else {
			checkDoor(level, Pos{p.X, p.Y - 1})
		}

	case Down:
		if canWalk(level, Pos{p.X, p.Y + 1}) {
			level.Player.Y++
		} else {
			checkDoor(level, Pos{p.X, p.Y + 1})
		}
	case Right:
		if canWalk(level, Pos{p.X + 1, p.Y}) {
			level.Player.X++
		} else {
			checkDoor(level, Pos{p.X + 1, p.Y})
		}
	case Left:
		if canWalk(level, Pos{p.X - 1, p.Y}) {
			level.Player.X--
		} else {
			checkDoor(level, Pos{p.X - 1, p.Y})
		}
	case CloseWindow:
		fmt.Println("Trying to close window")
		fmt.Println(len(game.LevelChans))
		close(input.LevelChannel)
		chanIndex := 0
		for i, c := range game.LevelChans {
			if c == input.LevelChannel {
				chanIndex = i
				break
			}
		}
		game.LevelChans = append(game.LevelChans[:chanIndex], game.LevelChans[chanIndex+1:]...)
		fmt.Println(len(game.LevelChans))
	}
}

func getNeighbors(level *Level, pos Pos) []Pos {
	neighbors := make([]Pos, 0, 4)
	left := Pos{pos.X - 1, pos.Y}
	right := Pos{pos.X + 1, pos.Y}
	up := Pos{pos.X, pos.Y - 1}
	down := Pos{pos.X, pos.Y + 1}

	if canWalk(level, right) {
		neighbors = append(neighbors, right)
	}
	if canWalk(level, left) {
		neighbors = append(neighbors, left)
	}
	if canWalk(level, up) {
		neighbors = append(neighbors, up)
	}
	if canWalk(level, down) {
		neighbors = append(neighbors, down)
	}

	return neighbors
}

func (game *Game) bfs(start Pos) {
	frontier := make([]Pos, 0, 8)
	frontier = append(frontier, start)
	visited := make(map[Pos]bool)
	visited[start] = true
	game.Level.Debug = visited

	for len(frontier) > 0 {
		current := frontier[0]
		frontier = frontier[1:]

		for _, next := range getNeighbors(game.Level, current) {
			if !visited[next] {
				frontier = append(frontier, next)
				visited[next] = true
				// ui.Draw(level)
				// time.Sleep(100 * time.Millisecond)
			}
		}
	}
}

func (game *Game) astar(start Pos, goal Pos) []Pos {
	frontier := make(pqueue, 0, 8)
	frontier = frontier.push(start, 1)
	cameFrom := make(map[Pos]Pos)
	cameFrom[start] = start
	costSoFar := make(map[Pos]int)
	costSoFar[start] = 0

	game.Level.Debug = make(map[Pos]bool)
	var current Pos
	// fmt.Println(len(frontier))
	for len(frontier) > 0 {

		frontier, current = frontier.pop()
		if current == goal { // Done with the search
			path := make([]Pos, 0)
			p := current
			for p != start {
				path = append(path, p)

				p = cameFrom[p]
			}
			path = append(path, p)
			// level.Debug[p] = true

			for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 { //Reversing the array
				path[i], path[j] = path[j], path[i]
			}

			game.Level.Debug = make(map[Pos]bool) // Draws the search now reversed
			for _, pos := range path {
				game.Level.Debug[pos] = true
				// ui.Draw(level)
				// time.Sleep(100 * time.Millisecond)
			}
			return path
		}

		for _, next := range getNeighbors(game.Level, current) { //Does the search
			newCost := costSoFar[current] + 1
			_, exists := costSoFar[next]
			if !exists || newCost < costSoFar[next] {
				costSoFar[next] = newCost
				xDist := int(math.Abs(float64(goal.X - next.X)))
				yDist := int(math.Abs(float64(goal.Y - next.Y)))
				priority := newCost + xDist + yDist
				frontier = frontier.push(next, priority)
				// level.Debug[next] = true
				// ui.Draw(level)
				// time.Sleep(100 * time.Millisecond)
				cameFrom[next] = current

			}
		}
	}

	return nil

}

func (game *Game) Run() {
	// level := loadLevelFromFile("game/maps/level1.map")

	for _, lchan := range game.LevelChans {
		lchan <- game.Level
	}

	for input := range game.InputChan {
		fmt.Println("Checking inputs")
		fmt.Println(input.Typ)
		if input.Typ == QuitGame {
			fmt.Println("Quitting")
			return
		}
		game.handleInput(input)

		if len(game.LevelChans) == 0 {
			return
		}

		// Send game state updates
		for _, lchan := range game.LevelChans {
			lchan <- game.Level
		}
	}

}
