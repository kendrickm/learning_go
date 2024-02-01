package game

import (
	"bufio"
	"os"
)

type GameUI interface {
	Draw(*Level)
}

type Tile rune

const (
	StoneWall Tile = '#'
	DirtFloor Tile = '.'
	Door      Tile = '|'
	Blank     Tile = ' '
)

type Level struct {
	Map [][]Tile
}

func Run(ui GameUI) {
	level := loadLevelFromFile("game/maps/level1.map")
	ui.Draw(level)
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
		//fmt.Println(scanner.Text())
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
		for x := 0; x < longestRow; x++ {
			for _, c := range line {
				var t Tile
				switch c {
				case ' ', '\n', '\t', '\r':
					t = Blank
				case '#':
					t = StoneWall
				case '|':
					t = Door
				case '.':
					t = DirtFloor
				default:
					panic("Invalid character in map")
				}
				level.Map[y][x] = t

			}
		}
	}
	return level
}
