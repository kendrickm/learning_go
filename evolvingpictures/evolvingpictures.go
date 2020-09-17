package main

import (
	"fmt"
	"math/rand"
	"time"

	. "github.com/kendrickm/learning_go/evolvingpictures/apt"
	. "github.com/kendrickm/learning_go/evolvingpictures/gui"

	// . "github.com/kendrickm/vec3"
	"github.com/veandco/go-sdl2/sdl"
)

var winWidth, winHeight int = 800, 600
var rows, cols, numPics int = 2, 2, rows * cols

type pixelResult struct {
	pixels []byte
	index  int
}

type audioState struct {
	explosionBytes []byte
	deviceID       sdl.AudioDeviceID
	audioSpec      *sdl.AudioSpec
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
	rand.Seed(time.Now().UTC().UnixNano())
	p.r = GetRandomNode()
	p.g = GetRandomNode()
	p.b = GetRandomNode()

	num := rand.Intn(20) + 5
	for i := 0; i < num; i++ {
		p.r.AddRandom(GetRandomNode())
	}

	num = rand.Intn(20) + 5
	for i := 0; i < num; i++ {
		p.g.AddRandom(GetRandomNode())
	}

	num = rand.Intn(20) + 5
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
	fmt.Println(nodeToMutate)

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

func aptToPixels(pic *picture, w, h int) []byte {
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
	return pixels
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

	rand.Seed(time.Now().UTC().UnixNano())

	picTrees := make([]*picture, numPics)
	for i := range picTrees {
		picTrees[i] = NewPicture()
		// fmt.Println(picTrees[i])
	}

	picWidth := int(float32(winWidth/cols) * float32(.9))
	picHeight := int(float32(winHeight/rows) * float32(.9))

	pixelsChannel := make(chan pixelResult, numPics)

	buttons := make([]*ImageButton, numPics)
	for i := range picTrees {
		go func(i int) {
			pixels := aptToPixels(picTrees[i], picWidth, picHeight)
			pixelsChannel <- pixelResult{pixels, i}
		}(i)
	}

	// Changd after EP 06 to address MacOSX
	// OSX requires that you consume events for windows to open and work properly
	keyboardState := sdl.GetKeyboardState()
	mouseState := GetMouseState()
	for {
		frameStart := time.Now()

		mouseState.Update()
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch e := event.(type) {
			case *sdl.QuitEvent:
				return
			case *sdl.TouchFingerEvent:
				if e.Type == sdl.FINGERDOWN {
					fmt.Println("Click")
					touchX := int(e.X * float32(winWidth))
					touchY := int(e.Y * float32(winHeight))
					mouseState.X = touchX
					mouseState.Y = touchY
					mouseState.LeftButton = true
				}
			}
		}

		if keyboardState[sdl.SCANCODE_ESCAPE] != 0 {
			return
		}

		select {
		case pixelsAndIndex, ok := <-pixelsChannel:
			if ok {
				tex := pixelsToTexture(renderer, pixelsAndIndex.pixels, picWidth, picHeight)
				xi := pixelsAndIndex.index % cols
				yi := (pixelsAndIndex.index - xi) / cols
				x := int32(xi * picWidth)
				y := int32(yi * picHeight)
				xPad := int32(float32(winWidth) * .1 / float32(cols+1))
				yPad := int32(float32(winHeight) * .1 / float32(rows+1))
				x += xPad * (int32(xi) + 1)
				y += yPad * (int32(yi) + 1)
				rect := &sdl.Rect{x, y, int32(picWidth), int32(picHeight)}
				button := NewImageButton(renderer, tex, *rect, sdl.Color{255, 255, 255, 0})
				buttons[pixelsAndIndex.index] = button
			}
		default:
		}

		renderer.Clear()
		for _, button := range buttons {
			if button != nil {
				button.Update(*mouseState)
				if button.WasLeftClicked {
					button.IsSelcted = !button.IsSelcted
				}
				button.Draw(renderer)
			}
		}

		renderer.Present()
		elapsedTime = float32(time.Since(frameStart).Seconds() * 1000)
		if elapsedTime < 5 {
			sdl.Delay(5 - uint32(elapsedTime))
			elapsedTime = float32(time.Since(frameStart).Seconds() * 1000)
		}
	}
}
