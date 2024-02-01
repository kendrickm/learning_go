package ui2d

import (
	"bufio"
	"fmt"
	"log"
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
var textureIndex map[game.Tile]sdl.Rect

func loadTextureIndex() {
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
		splitXy := strings.Split(xy, ",")
		x, err := strconv.ParseInt(strings.TrimSpace(splitXy[0]), 10, 64)
		if err != nil {
			panic(err)
		}
		y, err := strconv.ParseInt(strings.TrimSpace(splitXy[1]), 10, 64)
		if err != nil {
			panic(err)
		}
		fmt.Println(tile, x, y)
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
}

type UI2d struct {
}

func (*UI2d) Draw(level *game.Level) {
	fmt.Println("We did something")

	renderer.Copy(textureAtlas, nil, nil)
	renderer.Present()
   for {
	for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
	
	  switch  event.(type) {
          case *sdl.QuitEvent:
             return
          }
        }
    }
	sdl.Delay(3000)
}
