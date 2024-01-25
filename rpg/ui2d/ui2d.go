package ui2d

import (
	"fmt"
	"unsafe"

	"github.com/kendrickm/learning_go/rpg/game"
	"github.com/veandco/go-sdl2/img"
	"github.com/veandco/go-sdl2/sdl"
)

func f(p unsafe.Pointer) {}

const winWidth, winHeight = 1280, 720

var renderer *sdl.Renderer
var textureAtlas *sdl.Texture

func imgFileToTexture(filename string) *sdl.Texture {
	image, err := img.Load(filename)
	if err != nil {
		panic(err)
	}

	// infile, err := os.Open(filename)
	// if err != nil {
	// 	panic(err)
	// }
	// defer infile.Close()

	// img, err := png.Decode(infile)
	// if err != nil {
	// 	panic(err)
	// }

	// w := img.Bounds().Max.X
	// h := img.Bounds().Max.Y
	// //var pixels unsafe.Pointer
	// pixels := make([]byte, w*h*4)
	// bIndex := 0

	// for y := 0; y < h; y++ {
	// 	for x := 0; x < w; x++ {
	// 		r, g, b, a := img.At(x, y).RGBA()
	// 		pixels[bIndex] = byte(r / 256)
	// 		bIndex++
	// 		pixels[bIndex] = byte(g / 256)
	// 		bIndex++
	// 		pixels[bIndex] = byte(b / 256)
	// 		bIndex++
	// 		pixels[bIndex] = byte(a / 256)
	// 		bIndex++
	// 	}
	// }

	tex, err := renderer.CreateTextureFromSurface(image)
	if err != nil {
		panic(err)
	}
	//tex.Update(nil, unsafe.Pointer(&pixels[0]), w*4)

	err = tex.SetBlendMode(sdl.BLENDMODE_BLEND)
	if err != nil {
		panic(err)
	}

	return tex
}

func init() {
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
}

type UI2d struct {
}

func (*UI2d) Draw(level *game.Level) {
	fmt.Println("We did something")
	renderer.Copy(textureAtlas, nil, nil)
	renderer.Present()
	sdl.Delay(5000)
}
