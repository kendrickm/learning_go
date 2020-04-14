package main

import (
	"fmt"
	"math/rand"
	"time"

	. "github.com/kendrickm/learning_go/evolvingpictures/apt"
	// . "github.com/kendrickm/vec3"
	"github.com/veandco/go-sdl2/sdl"
)

const winWidth, winHeight, winDepth int = 800, 600, 100

type audioState struct {
	explosionBytes []byte
	deviceID       sdl.AudioDeviceID
	audioSpec      *sdl.AudioSpec
}

type mouseState struct {
	leftButton  bool
	rightButton bool
	x, y        int
}

func getMouseState() mouseState {
	mouseX, mouseY, mouseButtonState := sdl.GetMouseState()
	leftButton := mouseButtonState & sdl.ButtonLMask()
	rightButton := mouseButtonState & sdl.ButtonRMask()

	var result mouseState
	result.x = int(mouseX)
	result.y = int(mouseY)
	result.leftButton = !(leftButton == 0)
	result.rightButton = !(rightButton == 0)

	return result
}

type rgba struct {
	r, g, b byte
}

type picture struct {
	r, g, b Node
}

func (p *picture) String() string {
	return "R" + p.r.String() + "\n" + "G" + p.g.String() + "\n" + "B" + p.b.String()
}

func NewPicture() *picture {
	p := &picture{}

	p.r = GetRandomNode()
	p.g = GetRandomNode()
	p.b = GetRandomNode()
	rand.Seed(time.Now().UTC().UnixNano())

	num := rand.Intn(4)
	for i := 0; i < num; i++ {
		p.r.AddRandom(GetRandomNode())
	}

	num = rand.Intn(4)
	for i := 0; i < num; i++ {
		p.g.AddRandom(GetRandomNode())
	}

	num = rand.Intn(4)
	for i := 0; i < num; i++ {
		p.b.AddRandom(GetRandomNode())
	}

	for p.r.AddLeaf(GetRandomLeafNode()) {

	}

	for p.g.AddLeaf(GetRandomLeafNode()) {

	}

	for p.b.AddLeaf(GetRandomLeafNode()) {

	}

	return p
}

func (p *picture) Mutate() {
	r := rand.Intn(3)
	var nodeToMutate Node
	switch r {
	case 0:
		nodeToMutate = p.r
	case 1:
		nodeToMutate = p.g
	case 2:
		nodeToMutate = p.b
	}

	count := nodeToMutate.NodeCount()
	r = rand.Intn(count)

	nodeToMutate, count = GetNthNode(nodeToMutate, r, 0)
	mutation := Mutate(nodeToMutate)

	if mutation == p.r {
		p.r = mutation
	} else if mutation == p.g {
		p.g = mutation
	} else if mutation == p.b {
		p.b = mutation
	}
}

func pixelsToTexture(renderer *sdl.Renderer, pixels []byte, w, h int) *sdl.Texture {
	tex, err := renderer.CreateTexture(sdl.PIXELFORMAT_ABGR8888, sdl.TEXTUREACCESS_STREAMING, int32(w), int32(h))
	if err != nil {
		panic(err)
	}
	tex.Update(nil, pixels, w*4)
	return tex
}

func aptToTexture(pic *picture, w, h int, renderer *sdl.Renderer) *sdl.Texture {
	// -1.0 to 1.0
	scale := float32(255 / 2)
	offset := float32(1.0 * scale)
	pixels := make([]byte, w*h*4)
	pixelIndex := 0
	for yi := 0; yi < h; yi++ {
		y := float32(yi)/float32(h)*2 - 1
		for xi := 0; xi < w; xi++ {
			x := float32(xi)/float32(w)*2 - 1

			r := pic.r.Eval(x, y)
			g := pic.g.Eval(x, y)
			b := pic.g.Eval(x, y)

			pixels[pixelIndex] = byte(r*scale - offset)
			pixelIndex++
			pixels[pixelIndex] = byte(g*scale - offset)
			pixelIndex++
			pixels[pixelIndex] = byte(b*scale - offset)
			pixelIndex++
			pixelIndex++ //Skipping Alpha
		}
	}
	return pixelsToTexture(renderer, pixels, w, h)
}

func main() {

	sdl.LogSetAllPriority(sdl.LOG_PRIORITY_VERBOSE)
	err := sdl.Init(sdl.INIT_EVERYTHING)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer sdl.Quit()

	window, err := sdl.CreateWindow("Evolving Pictures", 200, 200,
		int32(winWidth), int32(winHeight), sdl.WINDOW_SHOWN)

	if err != nil {
		fmt.Println(err)
		return
	}
	defer window.Destroy()

	renderer, err := sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer renderer.Destroy()
	sdl.SetHint(sdl.HINT_RENDER_SCALE_QUALITY, "1")

	var elapsedTime float32
	currentMouseState := getMouseState()
	prevMouseState := currentMouseState

	rand.Seed(time.Now().UTC().UnixNano())

	pic := NewPicture()
	tex := aptToTexture(pic, 800, 600, renderer)

	// Changd after EP 06 to address MacOSX
	// OSX requires that you consume events for windows to open and work properly
	for {
		frameStart := time.Now()

		currentMouseState = getMouseState()

		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch e := event.(type) {
			case *sdl.QuitEvent:
				return
			case *sdl.TouchFingerEvent:
				if e.Type == sdl.FINGERDOWN {
					touchX := int(e.X * float32(winWidth))
					touchY := int(e.Y * float32(winHeight))
					currentMouseState.x = touchX
					currentMouseState.y = touchY
					currentMouseState.leftButton = true
				}
			}
		}

		if prevMouseState.leftButton && !currentMouseState.leftButton {
			pic.Mutate()
			tex = aptToTexture(pic, 800, 600, renderer)
		}

		renderer.Copy(tex, nil, nil)

		renderer.Present()
		elapsedTime = float32(time.Since(frameStart).Seconds() * 1000)
		if elapsedTime < 5 {
			sdl.Delay(5 - uint32(elapsedTime))
			elapsedTime = float32(time.Since(frameStart).Seconds() * 1000)
		}

		prevMouseState = currentMouseState
	}
}
