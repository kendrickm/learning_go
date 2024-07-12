package game

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"strconv"
)

type Game struct {
	LevelChans []chan *Level
	InputChan  chan *Input
	Level      *Level
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

type Tile struct {
	Rune    rune
	Visible bool
	//visited bool
}

const (
	StoneWall  rune = '#'
	DirtFloor  rune = '.'
	ClosedDoor rune = '|'
	OpenDoor   rune = '/'
	Blank      rune = 0
	Pending    rune = -1
)

type Pos struct {
	X, Y int
}

type Entity struct {
	Pos
	Name string
	Rune rune
}

type Character struct {
	Entity
	Hitpoints    int
	Strength     int
	Speed        float64
	ActionPoints float64
	SightRange   int
}

type Player struct {
	Character
}

type Level struct {
	Map      [][]Tile
	Player   *Player
	Monsters map[Pos]*Monster
	Debug    map[Pos]bool
	Events   []string
	EventPos int
}

func (level *Level) Attack(c1, c2 *Character) {
	c1.ActionPoints -= 1
	c2.Hitpoints -= c1.Strength

	if c2.Hitpoints > 0 {
		level.AddEvent(c1.Name + " Attacked " + c2.Name + " for " + strconv.Itoa(c1.Strength))
	} else {
		level.AddEvent(c1.Name + " Killed " + c2.Name)
	}
}

func (level *Level) AddEvent(event string) {
	level.Events[level.EventPos] = event
	level.EventPos++
	if level.EventPos == len(level.Events) {
		level.EventPos = 0
	}
}

func (level *Level) lineOfSight() {
	pos := level.Player.Pos
	dist := level.Player.SightRange

	for y := pos.Y - dist; y <= pos.Y+dist; y++ {
		for x := pos.X - dist; x <= pos.X + dist; x++ {
			xDelta := pos.X - x
			yDelta := pos.Y - y
			d := math.Sqrt(float64(xDelta*xDelta + yDelta*yDelta))
			if d <= float64(dist) {
				level.bresenham(pos,Pos{x,y})
			}
		} 
	}
}

func (level *Level) bresenham(start Pos, end Pos) {

	steep := math.Abs(float64(end.Y-start.Y)) > math.Abs(float64(end.X-start.X))
	if steep {
		start.X, start.Y = start.Y, start.X
		end.X, end.Y = end.Y, end.X
	}

	deltaY := int(math.Abs(float64(end.Y - start.Y)))
	err := 0
	y := start.Y
	ystep := 1
	if start.Y >= end.Y {
		ystep = -1
	}

	if start.X > end.X {
		deltaX := start.X - end.X		
		for x := start.X; x >= end.X; x-- {
			var pos Pos
			if steep {
				pos = Pos{y,x}
			} else {
				pos = Pos{x,y}
			}
			level.Map[pos.Y][pos.X].Visible = true
			if !canSeeThrough(level,pos) {
				return
			}
			err += deltaY
			if 2*err >= deltaX {
				y += ystep
				err -= deltaX
			}
		}
	} else {
		deltaX := end.X - start.X
		for x := start.X; x < end.X; x++ {
			var pos Pos
			if steep {
				pos = Pos{y,x}
			} else {
				pos = Pos{x,y}
			}
			level.Map[pos.Y][pos.X].Visible = true
			if !canSeeThrough(level,pos) {
				return
			}
			err += deltaY
			if 2*err >= deltaX {
				y += ystep
				err -= deltaX
			}
		}
	}

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
	level.Debug = make(map[Pos]bool)
	level.Events = make([]string, 10)
	level.Player = &Player{}
	level.Player.Strength = 1
	level.Player.Hitpoints = 20
	level.Player.Name = "Go"
	level.Player.Rune = '@'
	level.Player.ActionPoints = 0.0
	level.Player.SightRange = 3

	level.Map = make([][]Tile, len(levelLines))
	level.Monsters = make(map[Pos]*Monster)
	for i := range level.Map {
		level.Map[i] = make([]Tile, longestRow)
	}

	for y := 0; y < len(level.Map); y++ {
		line := levelLines[y]

		for x, c := range line {
			var t Tile

			switch c {
			case ' ', '\n', '\t', '\r':
				t.Rune = Blank
			case '#':
				t.Rune = StoneWall
			case '|':
				t.Rune = ClosedDoor
			case '/':
				t.Rune = OpenDoor
			case '.':
				t.Rune = DirtFloor
			case '@':
				level.Player.X = x
				level.Player.Y = y
				t.Rune = Pending //The tile under this is filled in later
			case 'R':
				level.Monsters[Pos{x, y}] = NewRat(Pos{x, y})
				t.Rune = Pending
			case 'S':
				level.Monsters[Pos{x, y}] = NewSpider(Pos{x, y})
				t.Rune = Pending
			default:
				panic("Invalid character in map")
			}
			level.Map[y][x] = t

		}
	}

	//TODO: Use BFS for this
	for y, row := range level.Map {
		for x, tile := range row {
			if tile.Rune == Pending {
				level.Map[y][x] = level.bfsFloor(Pos{x, y})
			}
		}
	}
	level.lineOfSight()
	return level
}
func inRange(level *Level, pos Pos) bool {
	return (pos.X < len(level.Map[0]) && pos.Y < len(level.Map) && pos.X >= 0 && pos.Y >= 0)
}

func canWalk(level *Level, pos Pos) bool {
	if inRange(level, pos) {
		t := level.Map[pos.Y][pos.X]
		switch t.Rune {
		case StoneWall, ClosedDoor, Blank:
			return false
		}
		_, exists := level.Monsters[pos]
		if exists {
			return false
		}
		return true
	}
	return false
}

func canSeeThrough(level *Level, pos Pos) bool {
	if inRange(level,pos) {
			t := level.Map[pos.Y][pos.X]
	switch t.Rune {
	case StoneWall, ClosedDoor, Blank:
		return false

	default:
		return true
	}
	}
	return false
}

func checkDoor(level *Level, pos Pos) {
	t := level.Map[pos.Y][pos.X]
	if t.Rune == ClosedDoor {
		level.Map[pos.Y][pos.X].Rune = OpenDoor
		level.lineOfSight()
	}
}

func (p *Player) Move(to Pos, level *Level) {
	p.Pos = to

	for y, row := range level.Map {
		for x, _ := range row {
			level.Map[y][x].Visible = false
		}
	}
	level.lineOfSight()
	//level.bresenham(p.Pos, Pos{p.Pos.X, p.Pos.Y - p.SightRange})
}

func (level *Level) resolveMovement(pos Pos) {
	monster, exists := level.Monsters[pos]
	if exists {
		level.Attack(&level.Player.Character, &monster.Character)
		if monster.Hitpoints <= 0 {
			delete(level.Monsters, monster.Pos)
		}
		if level.Player.Hitpoints <= 0 {
			panic("you ded")
		}
	} else if canWalk(level, pos) {
		level.Player.Move(pos, level)
	} else {
		checkDoor(level, pos)
	}

}

func (game *Game) handleInput(input *Input) {
	level := game.Level
	p := level.Player
	switch input.Typ {
	case Up:
		newPos := Pos{p.X, p.Y - 1}
		level.resolveMovement(newPos)
	case Down:
		newPos := Pos{p.X, p.Y + 1}
		level.resolveMovement(newPos)
	case Right:
		newPos := Pos{p.X + 1, p.Y}
		level.resolveMovement(newPos)
	case Left:
		newPos := Pos{p.X - 1, p.Y}
		level.resolveMovement(newPos)
	case CloseWindow:
		close(input.LevelChannel)
		chanIndex := 0
		for i, c := range game.LevelChans {
			if c == input.LevelChannel {
				chanIndex = i
				break
			}
		}
		game.LevelChans = append(game.LevelChans[:chanIndex], game.LevelChans[chanIndex+1:]...)
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

func (level *Level) bfsFloor(start Pos) Tile {
	frontier := make([]Pos, 0, 8)
	frontier = append(frontier, start)
	visited := make(map[Pos]bool)
	visited[start] = true

	for len(frontier) > 0 {
		current := frontier[0]
		currentTile := level.Map[current.Y][current.X]
		switch currentTile.Rune {
		case DirtFloor:
			return Tile{DirtFloor, false}
		default:

		}

		frontier = frontier[1:]

		for _, next := range getNeighbors(level, current) {
			if !visited[next] {
				frontier = append(frontier, next)
				visited[next] = true
			}
		}
	}

	return Tile{DirtFloor, false}
}

func (level *Level) astar(start Pos, goal Pos) []Pos {
	frontier := make(pqueue, 0, 8)
	frontier = frontier.push(start, 1)
	cameFrom := make(map[Pos]Pos)
	cameFrom[start] = start
	costSoFar := make(map[Pos]int)
	costSoFar[start] = 0

	var current Pos

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

			for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 { //Reversing the array
				path[i], path[j] = path[j], path[i]
			}

			return path
		}

		for _, next := range getNeighbors(level, current) { //Does the search
			newCost := costSoFar[current] + 1
			_, exists := costSoFar[next]
			if !exists || newCost < costSoFar[next] {
				costSoFar[next] = newCost
				xDist := int(math.Abs(float64(goal.X - next.X)))
				yDist := int(math.Abs(float64(goal.Y - next.Y)))
				priority := newCost + xDist + yDist
				frontier = frontier.push(next, priority)
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
		fmt.Println(input.Typ)
		if input.Typ == QuitGame {
			fmt.Println("Quitting")
			return
		}

		game.handleInput(input)
		for _, monster := range game.Level.Monsters {
			monster.Update(game.Level)
		}

		if len(game.LevelChans) == 0 {
			return
		}

		// Send game state updates
		for _, lchan := range game.LevelChans {
			lchan <- game.Level
		}
	}

}
