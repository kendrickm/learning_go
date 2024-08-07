package game

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Game struct {
	LevelChans   []chan *Level
	InputChan    chan *Input
	Levels       map[string]*Level
	CurrentLevel *Level
}

func NewGame(numWindows int) *Game {
	levelChans := make([]chan *Level, numWindows)
	for i := range levelChans {
		levelChans[i] = make(chan *Level)
	}

	inputChan := make(chan *Input)
	levels := loadLevels()
	game := &Game{levelChans, inputChan, levels, nil}
	game.loadWorldFile()
	game.CurrentLevel.lineOfSight()
	return game

}

type InputType int

const (
	None InputType = iota
	Up
	Down
	Left
	Right
	TakeAll
	QuitGame
	CloseWindow
	TakeItem
	DropItem
	EquipItem
	Search // Temp
)

type Input struct {
	Typ          InputType
	Item 		 *Item
	LevelChannel chan *Level
}

type Tile struct {
	Rune        rune
	OverlayRune rune
	Visible     bool
	Seen        bool
}

const (
	StoneWall  rune = '#'
	DirtFloor  rune = '.'
	ClosedDoor rune = '|'
	OpenDoor   rune = '/'
	UpStair    rune = 'u'
	DownStair  rune = 'd'
	Blank      rune = 0
	Pending    rune = -1
)

type Pos struct {
	X, Y int
}

type LevelPos struct {
	*Level
	Pos
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
	Items []*Item
	Helmet *Item
	Weapon *Item
}

type Player struct {
	Character
}

type GameEvent int
const(
	Move GameEvent = iota
	DoorOpen 
	Attack
	Hit
	Portal
	Pickup
	Drop
)

type Level struct {
	Map      [][]Tile
	Player   *Player
	Monsters map[Pos]*Monster
	Items    map[Pos][]*Item
	Portal   map[Pos]*LevelPos
	Debug    map[Pos]bool
	Events   []string
	EventPos int
	LastEvent GameEvent
}

func (level *Level) DropItem(itemToDrop *Item, character *Character) {
	pos := character.Pos
	items := character.Items
	for i,item := range items {
		if item == itemToDrop {
			character.Items = append(character.Items[:i],character.Items[i+1:]...)
			level.Items[pos] = append(level.Items[pos], item)
			level.AddEvent(character.Name + " dropped: " + item.Name)
			return
		}
	}
}

func (level *Level) MoveItem(itemToMove *Item, character *Character) {
	fmt.Println("Move Item")
	pos := character.Pos
	items := level.Items[pos]
	for i,item := range items {
		if item == itemToMove {
			items = append(items[:i],items[i+1:]...)
			level.Items[pos] = items
			character.Items = append(character.Items, item)
			level.AddEvent(character.Name + " picked up: " + item.Name)
			return
		}
	}
	panic ("Tried to move an item we were not on top of")
}

func (level *Level) Attack(c1, c2 *Character) {
	c1.ActionPoints -= 1
	c1AttackPower := c1.Strength
	if c1.Weapon != nil {
		c1AttackPower = int(float64(c1AttackPower) * c1.Weapon.power)
	}

	damage := c1AttackPower

	if c2.Helmet != nil {
		damage = int(float64(damage) * (1-c2.Helmet.power))
	}

	c2.Hitpoints -= damage

	if c2.Hitpoints > 0 {
		level.AddEvent(c1.Name + " Attacked " + c2.Name + " for " + strconv.Itoa(damage))
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
		for x := pos.X - dist; x <= pos.X+dist; x++ {
			xDelta := pos.X - x
			yDelta := pos.Y - y
			d := math.Sqrt(float64(xDelta*xDelta + yDelta*yDelta))
			if d <= float64(dist) {
				level.bresenham(pos, Pos{x, y})
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
		for x := start.X; x > end.X; x-- {
			var pos Pos
			if steep {
				pos = Pos{y, x}
			} else {
				pos = Pos{x, y}
			}
			level.Map[pos.Y][pos.X].Visible = true
			level.Map[pos.Y][pos.X].Seen = true
			if !canSeeThrough(level, pos) {
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
				pos = Pos{y, x}
			} else {
				pos = Pos{x, y}
			}
			level.Map[pos.Y][pos.X].Visible = true
			level.Map[pos.Y][pos.X].Seen = true
			if !canSeeThrough(level, pos) {
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

func (game *Game) loadWorldFile() {
	file, err := os.Open("game/maps/world.txt")
	if err != nil {
		panic(err)
	}

	csvReader := csv.NewReader(file)
	csvReader.FieldsPerRecord = -1
	csvReader.TrimLeadingSpace = true

	rows, err := csvReader.ReadAll()
	if err != nil {
		panic(err)
	}

	for rowIndex, row := range rows {
		//Set current level
		if rowIndex == 0 {
			game.CurrentLevel = game.Levels[row[0]]
			if game.CurrentLevel == nil {
				fmt.Println("Coudln't find currentlevel name in world file", row[0])
				panic(nil)
			}
			continue
		}
		levelWithPortal := game.Levels[row[0]]
		if levelWithPortal == nil {
			fmt.Println("Coudln't find level name 1 in world file", row[0])
			panic(nil)
		}
		x, err := strconv.ParseInt(row[1], 10, 64)
		if err != nil {
			panic(err)
		}
		y, err := strconv.ParseInt(row[2], 10, 64)
		if err != nil {
			panic(err)
		}
		pos := Pos{int(x), int(y)}

		levelToTeleportTo := game.Levels[row[3]]
		if levelToTeleportTo == nil {
			fmt.Println("Coudln't find level name 2 in world file", row[3])
			panic(nil)
		}

		x, err = strconv.ParseInt(row[4], 10, 64)
		if err != nil {
			panic(err)
		}
		y, err = strconv.ParseInt(row[5], 10, 64)
		if err != nil {
			panic(err)
		}
		posToTeleportTo := Pos{int(x), int(y)}

		levelWithPortal.Portal[pos] = &LevelPos{levelToTeleportTo, posToTeleportTo}
	}
}

// TODO take in path
func loadLevels() map[string]*Level {
	// TODO refactor the depreacted function

	filenames, err := filepath.Glob("game/maps/*.map")
	if err != nil {
		panic(err)
	}

	player := &Player{}
	player.Strength = 1
	player.Hitpoints = 100
	player.Name = "Go"
	player.Rune = '@'
	player.ActionPoints = 0.0
	player.SightRange = 7

	levels := make(map[string]*Level)

	for _, filename := range filenames {

		//extIndex := strings.LastIndex(filename, ".map")
		// lastSlashIndex := strings.LastIndex(filename, "\\")
		// lastSlashIndex := strings.LastIndex(filename, "/")
		//levelName := filepath.Base(filename)
		// levelName := filename[lastSlashIndex+1 : extIndex]
		levelName := strings.TrimSuffix(filepath.Base(filename), filepath.Ext(filename))
		fmt.Println("loading:", levelName)
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
		level.Player = player

		level.Map = make([][]Tile, len(levelLines))
		level.Monsters = make(map[Pos]*Monster)
		level.Items = make(map[Pos][]*Item)
		level.Portal = make(map[Pos]*LevelPos)

		for i := range level.Map {
			level.Map[i] = make([]Tile, longestRow)
		}

		for y := 0; y < len(level.Map); y++ {
			line := levelLines[y]

			for x, c := range line {
				var t Tile
				p := Pos{x,y}
				t.OverlayRune = Blank
				switch c {
				case ' ', '\n', '\t', '\r':
					t.Rune = Blank
				case '#':
					t.Rune = StoneWall
				case '|':
					t.OverlayRune = ClosedDoor
					t.Rune = Pending
				case '/':
					t.Rune = Pending
					t.OverlayRune = OpenDoor
				case 'u':
					t.Rune = Pending
					t.OverlayRune = UpStair
				case 'd':
					t.Rune = Pending
					t.OverlayRune = DownStair
				case 's':
					t.Rune = Pending
					level.Items[p] = append(level.Items[p], NewSword(p))
					level.Items[p] = append(level.Items[p], NewHelmet(p))
				case 'h':
					t.Rune = Pending
					level.Items[p] = append(level.Items[p], NewHelmet(p))
				case '.':
					t.Rune = DirtFloor
				case '@':
					level.Player.X = x
					level.Player.Y = y
					t.Rune = Pending //The tile under this is filled in later
				case 'R':
					level.Monsters[p] = NewRat(p)
					t.Rune = Pending
				case 'S':
					level.Monsters[p] = NewSpider(p)
					t.Rune = Pending
				default:
					panic("Invalid character in map")
				}
				level.Map[y][x] = t

			}
		}

		for y, row := range level.Map {
			for x, tile := range row {
				if tile.Rune == Pending {
					level.Map[y][x].Rune = level.bfsFloor(Pos{x, y})
				}
			}
		}
		levels[levelName] = level
	}
	return levels
}
func inRange(level *Level, pos Pos) bool {
	return (pos.X < len(level.Map[0]) && pos.Y < len(level.Map) && pos.X >= 0 && pos.Y >= 0)
}

func canWalk(level *Level, pos Pos) bool {
	if inRange(level, pos) {
		t := level.Map[pos.Y][pos.X]
		switch t.Rune {
		case StoneWall, Blank:
			return false
		}
		switch t.OverlayRune {
		case ClosedDoor:
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
	if inRange(level, pos) {
		t := level.Map[pos.Y][pos.X]
		switch t.Rune {
		case StoneWall, Blank:
			return false

		}

		switch t.OverlayRune {
		case ClosedDoor:
			return false
		default:
			return true
		}
	}

	return false
}

func checkDoor(level *Level, pos Pos) {
	t := level.Map[pos.Y][pos.X]
	if t.OverlayRune == ClosedDoor {
		level.Map[pos.Y][pos.X].OverlayRune = OpenDoor
		level.LastEvent = DoorOpen
		level.lineOfSight()
	}
}

func (game *Game) Move(to Pos) {
	level := game.CurrentLevel
	levelAndPos := level.Portal[to]
	player := game.CurrentLevel.Player
	fmt.Println(to)
	if levelAndPos != nil {
		game.CurrentLevel = levelAndPos.Level
		game.CurrentLevel.Player.Pos = levelAndPos.Pos
		game.CurrentLevel.lineOfSight()
	} else {

		player.Pos = to
		// items := level.Items[player.Pos]
		// if len(items) > 0 {
		// 	level.MoveItem(items[0], &player.Character)
		// 	fmt.Println("Player Inventory")
		// 	for _, item := range player.Items {
		// 		fmt.Println("	",item)
		// 	}
		// }
		level.LastEvent = Move
		for y, row := range level.Map {
			for x, _ := range row {
				level.Map[y][x].Visible = false
			}
		}
		level.lineOfSight()
	}
}

func (game *Game) resolveMovement(pos Pos) {
	level := game.CurrentLevel
	monster, exists := level.Monsters[pos]
	if exists {
		level.Attack(&level.Player.Character, &monster.Character)
		level.LastEvent = Attack
		if monster.Hitpoints <= 0 {
			monster.Kill(level)
			
		}
		if level.Player.Hitpoints <= 0 {
			panic("you ded")
		}
	} else if canWalk(level, pos) {
		game.Move(pos)
	} else {
		checkDoor(level, pos)
	}

}

func equip(c *Character, itemtoEquip *Item) {
	for i,item := range c.Items {
		if item == itemtoEquip { // Remove from inventory and then equip
			c.Items = append(c.Items[:i], c.Items[i+1:]...)
			if itemtoEquip.Typ == Helmet {
		 		c.Helmet = itemtoEquip
		 	} else if itemtoEquip.Typ == Weapon {
		 		c.Weapon = itemtoEquip
		 	
		 }
		 return
		}
	}

	panic("Someone tried to equip a thing they don't have")
}

func (game *Game) handleInput(input *Input) {
	level := game.CurrentLevel
	p := level.Player
	switch input.Typ {
	case Up:
		newPos := Pos{p.X, p.Y - 1}
		game.resolveMovement(newPos)
	case Down:
		newPos := Pos{p.X, p.Y + 1}
		game.resolveMovement(newPos)
	case Right:
		newPos := Pos{p.X + 1, p.Y}
		game.resolveMovement(newPos)
	case Left:
		newPos := Pos{p.X - 1, p.Y}
		game.resolveMovement(newPos)
	case TakeItem:
		level.MoveItem(input.Item, &level.Player.Character)
		level.LastEvent = Pickup
	case DropItem:
		level.DropItem(input.Item, &level.Player.Character)
		level.LastEvent = Drop
	case EquipItem:
		equip(&level.Player.Character, input.Item)
	case TakeAll:
		for _, item := range level.Items[p.Pos] {
			level.MoveItem(item, &p.Character)
		}
		level.LastEvent = Pickup
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

func (level *Level) bfsFloor(start Pos) rune {
	frontier := make([]Pos, 0, 8)
	frontier = append(frontier, start)
	visited := make(map[Pos]bool)
	visited[start] = true

	for len(frontier) > 0 {
		current := frontier[0]
		currentTile := level.Map[current.Y][current.X]
		switch currentTile.Rune {
		case DirtFloor:
			return DirtFloor
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

	return DirtFloor
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
		lchan <- game.CurrentLevel
	}

	for input := range game.InputChan {
		fmt.Println(input.Typ)
		if input.Typ == QuitGame {
			fmt.Println("Quitting")
			return
		}

		game.handleInput(input)
		for _, monster := range game.CurrentLevel.Monsters {
			monster.Update(game.CurrentLevel)
		}

		if len(game.LevelChans) == 0 {
			return
		}

		// Send game state updates
		for _, lchan := range game.LevelChans {
			lchan <- game.CurrentLevel
		}
	}

}
