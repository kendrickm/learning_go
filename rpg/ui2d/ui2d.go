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

const winWidth, winHeight = 1280, 720

var renderer *sdl.Renderer
var textureAtlas *sdl.Texture
var textureIndex map[game.Tile][]sdl.Rect
var preKeyboardState []uint8
var keyboardState []uint8

var centerX int
var centerY int

func loadTextureIndex() {
	textureIndex = make(map[game.Tile][]sdl.Rect)
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

		textureIndex[tile] = rects
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

}

func imgFileToTexture(filename string) *sdl.Texture {
	image, err := img.Load(filename)
	if err != nil {
		panic(err)
	}

	tex, err := renderer.CreateTextureFromSurface(image)
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
	sdl.LogSetAllPriority(sdl.LOG_PRIORITY_VERBOSE)
	err := sdl.Init(sdl.INIT_EVERYTHING)
	if err != nil {
		fmt.Println(err)
		return
	}

	window, err := sdl.CreateWindow("RPG", 200, 200,
		int32(winWidth), int32(winHeight), sdl.WINDOW_SHOWN)

	if err != nil {
		fmt.Println(err)
		return
	}

	renderer, err = sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED)
	if err != nil {
		fmt.Println(err)
		return
	}
	sdl.SetHint(sdl.HINT_RENDER_SCALE_QUALITY, "1")

	textureAtlas = imgFileToTexture("ui2d/assets/tiles.png")
	loadTextureIndex()

	keyboardState = sdl.GetKeyboardState()
	preKeyboardState = make([]uint8, len(keyboardState))

	centerX = -1
	centerY = -1
}

type UI2d struct {
}

func (*UI2d) GetInput() *game.Input {

	for {
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {

			switch event.(type) {
			case *sdl.QuitEvent:
				return &game.Input{Typ: game.Quit}

			}
		}
		var input game.Input
		if keyboardState[sdl.SCANCODE_UP] == 0 && preKeyboardState[sdl.SCANCODE_UP] != 0 {
			input.Typ = game.Up
		}

		if keyboardState[sdl.SCANCODE_DOWN] == 0 && preKeyboardState[sdl.SCANCODE_DOWN] != 0 {
			input.Typ = game.Down
		}

		if keyboardState[sdl.SCANCODE_RIGHT] == 0 && preKeyboardState[sdl.SCANCODE_RIGHT] != 0 {
			input.Typ = game.Right
		}

		if keyboardState[sdl.SCANCODE_LEFT] == 0 && preKeyboardState[sdl.SCANCODE_LEFT] != 0 {
			input.Typ = game.Left
		}

		for i, v := range keyboardState {
			preKeyboardState[i] = v
		}

		if input.Typ != game.None {
			return &input
		}
	}

}

func (*UI2d) Draw(level *game.Level) {
	if centerX == -1 && centerY == -1 {
		centerX = level.Player.X
		centerY = level.Player.Y
	}

	limit := 5
	if level.Player.X > centerX+limit {
		centerX++
	} else if level.Player.X < centerX-limit {
		centerX--
	} else if level.Player.Y > centerY+limit {
		centerY++
	} else if level.Player.Y < centerY-limit {
		centerY--
	}

	offsetX := int32((winWidth / 2) - centerX*32)
	offsetY := int32((winHeight / 2) - centerY*32)
	renderer.Clear()
	rand.Seed(2)
	for y, row := range level.Map {
		for x, tile := range row {
			if tile != game.Blank {
				srcRects := textureIndex[tile]
				srcRect := srcRects[rand.Intn(len(srcRects))]
				dstRect := sdl.Rect{int32(x*32) + offsetX, int32(y*32) + offsetY, 32, 32}
				renderer.Copy(textureAtlas, &srcRect, &dstRect)
			}
			//fmt.Println(tile)
		}
	}

	renderer.Copy(textureAtlas, &sdl.Rect{21 * 32, 59 * 32, 32, 32}, &sdl.Rect{X: int32(level.Player.X)*32 + offsetX, Y: int32(level.Player.Y)*32 + offsetY, W: 32, H: 32})
	renderer.Present()

}
