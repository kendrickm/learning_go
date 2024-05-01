package ui2d

import (
	"bufio"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"

	"github.com/kendrickm/learning_go/rpg/game"
	"github.com/veandco/go-sdl2/img"
	"github.com/veandco/go-sdl2/sdl"
)

//func f(p unsafe.Pointer) {}

type ui struct {
	winWidth  int
	winHeight int

	renderer         *sdl.Renderer
	window           *sdl.Window
	textureAtlas     *sdl.Texture
	textureIndex     map[game.Tile][]sdl.Rect
	preKeyboardState []uint8
	keyboardState    []uint8
	r                *rand.Rand
	centerX          int
	centerY          int

	levelChan chan *game.Level
	inputChan chan *game.Input
}

func NewUI(inputChan chan *game.Input, levelChan chan *game.Level) *ui {
	ui := &ui{}
	ui.inputChan = inputChan
	ui.levelChan = levelChan
	ui.winHeight = 720
	ui.winWidth = 1080
	ui.r = rand.New(rand.NewSource(1))
	window, err := sdl.CreateWindow("RPG", 200, 200,
		int32(ui.winWidth), int32(ui.winHeight), sdl.WINDOW_SHOWN)
	if err != nil {
		panic(err)
	}
	ui.window = window

	ui.renderer, err = sdl.CreateRenderer(ui.window, -1, sdl.RENDERER_ACCELERATED)
	if err != nil {
		panic(err)
	}
	sdl.SetHint(sdl.HINT_RENDER_SCALE_QUALITY, "1")

	ui.textureAtlas = ui.imgFileToTexture("ui2d/assets/tiles.png")
	ui.loadTextureIndex()

	ui.keyboardState = sdl.GetKeyboardState()
	ui.preKeyboardState = make([]uint8, len(ui.keyboardState))
	for i, v := range ui.keyboardState {
		ui.preKeyboardState[i] = v
	}

	ui.centerX = -1
	ui.centerY = -1

	return ui
}

func (ui *ui) loadTextureIndex() {
	ui.textureIndex = make(map[game.Tile][]sdl.Rect)
	file, err := os.Open("ui2d/assets/asset-index.txt")
	if err != nil {
		panic(err)
	}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)
		tile := game.Tile(line[0])
		xy := line[1:]
		splitXYCount := strings.Split(xy, ",")
		x, err := strconv.ParseInt(strings.TrimSpace(splitXYCount[0]), 10, 64)
		if err != nil {
			panic(err)
		}
		y, err := strconv.ParseInt(strings.TrimSpace(splitXYCount[1]), 10, 64)
		if err != nil {
			panic(err)
		}
		variationCount, err := strconv.ParseInt(strings.TrimSpace(splitXYCount[2]), 10, 64) //Supports randomly picking from a batch of tiles
		if err != nil {
			panic(err)
		}
		var rects []sdl.Rect
		for i := int64(0); i < variationCount; i++ {
			rects = append(rects, sdl.Rect{int32(x * 32), int32(y * 32), 32, 32})
			x++
			if x > 62 { //handles wrap arounds
				x = 0
				y++
			}

		}

		ui.textureIndex[tile] = rects
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

}

func (ui *ui) imgFileToTexture(filename string) *sdl.Texture {
	image, err := img.Load(filename)
	if err != nil {
		panic(err)
	}

	tex, err := ui.renderer.CreateTextureFromSurface(image)
	if err != nil {
		panic(err)
	}

	err = tex.SetBlendMode(sdl.BLENDMODE_BLEND)
	if err != nil {
		panic(err)
	}

	return tex
}

func init() {
	fmt.Println("Init innit")
	err := sdl.Init(sdl.INIT_EVERYTHING)
	if err != nil {
		fmt.Println(err)
		return
	}
}

func (ui *ui) Run() {

	for {
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {

			switch e := event.(type) {
			case *sdl.QuitEvent:
				ui.inputChan <- &game.Input{Typ: game.QuitGame}
			case *sdl.WindowEvent:
				if e.Event == sdl.WINDOWEVENT_CLOSE {
					ui.inputChan <- &game.Input{Typ: game.CloseWindow, LevelChannel: ui.levelChan}
					return
				}

			}
		}
		select {
		case newLevel, ok := <-ui.levelChan:
			if ok {
				ui.Draw(newLevel)
			}
		default:
		}

		if sdl.GetKeyboardFocus() == ui.window || sdl.GetMouseFocus() == ui.window {
			var input game.Input
			if ui.keyboardState[sdl.SCANCODE_UP] == 0 && ui.preKeyboardState[sdl.SCANCODE_UP] != 0 {
				input.Typ = game.Up
			}

			if ui.keyboardState[sdl.SCANCODE_DOWN] == 0 && ui.preKeyboardState[sdl.SCANCODE_DOWN] != 0 {
				input.Typ = game.Down
			}

			if ui.keyboardState[sdl.SCANCODE_RIGHT] == 0 && ui.preKeyboardState[sdl.SCANCODE_RIGHT] != 0 {
				input.Typ = game.Right
			}

			if ui.keyboardState[sdl.SCANCODE_LEFT] == 0 && ui.preKeyboardState[sdl.SCANCODE_LEFT] != 0 {
				input.Typ = game.Left
			}

			if ui.keyboardState[sdl.SCANCODE_S] == 0 && ui.preKeyboardState[sdl.SCANCODE_S] != 0 {
				input.Typ = game.Search
			}

			for i, v := range ui.keyboardState {
				ui.preKeyboardState[i] = v
			}

			if input.Typ != game.None {
				ui.inputChan <- &input
			}
		}
		sdl.Delay(10)
	}

}

func (ui *ui) Draw(level *game.Level) {
	if ui.centerX == -1 && ui.centerY == -1 {
		ui.centerX = level.Player.X
		ui.centerY = level.Player.Y
	}

	limit := 5
	if level.Player.X > ui.centerX+limit {
		ui.centerX++
	} else if level.Player.X < ui.centerX-limit {
		ui.centerX--
	} else if level.Player.Y > ui.centerY+limit {
		ui.centerY++
	} else if level.Player.Y < ui.centerY-limit {
		ui.centerY--
	}

	offsetX := int32((ui.winWidth / 2) - ui.centerX*32)
	offsetY := int32((ui.winHeight / 2) - ui.centerY*32)
	ui.renderer.Clear()
	ui.r.Seed(2)
	for y, row := range level.Map {
		for x, tile := range row {
			if tile != game.Blank {
				srcRects := ui.textureIndex[tile]
				srcRect := srcRects[ui.r.Intn(len(srcRects))]
				dstRect := sdl.Rect{int32(x*32) + offsetX, int32(y*32) + offsetY, 32, 32}

				pos := game.Pos{x, y}
				if level.Debug[pos] {
					ui.textureAtlas.SetColorMod(128, 0, 0)
				} else {
					ui.textureAtlas.SetColorMod(255, 255, 255)
				}
				ui.renderer.Copy(ui.textureAtlas, &srcRect, &dstRect)
			}
			//fmt.Println(tile)
		}
	}

	for pos, monster := range level.Monsters {
		monsterSrcRect := ui.textureIndex[game.Tile(monster.Rune)][0]
		ui.renderer.Copy(ui.textureAtlas, &monsterSrcRect, &sdl.Rect{X: int32(pos.X)*32 + offsetX, Y: int32(pos.Y)*32 + offsetY, W: 32, H: 32})

	}
	playerSrcRect := ui.textureIndex['@'][0]
	ui.renderer.Copy(ui.textureAtlas, &playerSrcRect, &sdl.Rect{X: int32(level.Player.X)*32 + offsetX, Y: int32(level.Player.Y)*32 + offsetY, W: 32, H: 32})
	ui.renderer.Present()

}
