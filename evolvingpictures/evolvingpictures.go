// Homework Ideas
//
// 1. Make large image load in goroutine, display loading indication while it's loading
// 2. Instead of passing x,y for each pixel, pass a single array for all pixels and evaluate the whole  array at once
// 		measure and compare performance
// 3. Make the string() functions output valid Go code, and make a program that will run that code and render it
//		Make template source that renders the output
// 4. Currently we have R G and B for each picture. Do a greyscale picture with just one node, or HSV picture
// 		which uses Hue, Satuation, Value and then convert HSV to RGB

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

var winWidth, winHeight int = 1700, 900
var rows, cols, numPics int = 5, 5, rows * cols

type pixelResult struct {
	pixels []byte
	index  int
}

type guiState struct {
	zoom    bool
	zoomImg *sdl.Texture
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
	p.r = GetRandomNode()
	p.g = GetRandomNode()
	p.b = GetRandomNode()

	num := rand.Intn(25) + 5
	for i := 0; i < num; i++ {
		p.r.AddRandom(GetRandomNode())
	}

	num = rand.Intn(25) + 5
	for i := 0; i < num; i++ {
		p.g.AddRandom(GetRandomNode())
	}

	num = rand.Intn(25) + 5
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

func (p *picture) pickRandomColor() Node {
	r := rand.Intn(3)
	switch r {
	case 0:
		return p.r
	case 1:
		return p.g
	case 2:
		return p.b
	default:
		panic("pickRandomColor failed")
	}
}

func cross(a *picture, b *picture) *picture {
	aCopy := &picture{CopyTree(a.r, nil), CopyTree(a.g, nil), CopyTree(a.b, nil)}
	aColor := aCopy.pickRandomColor()
	bColor := b.pickRandomColor()

	aIndex := rand.Intn(aColor.NodeCount())
	aNode, _ := GetNthNode(aColor, aIndex, 0)

	bIndex := rand.Intn(bColor.NodeCount())
	bNode, _ := GetNthNode(bColor, bIndex, 0)
	bNodeCopy := CopyTree(bNode, bNode.GetParent())

	ReplaceNode(aNode, bNodeCopy)
	return aCopy
}

func evolve(survivors []*picture) []*picture {
	newPics := make([]*picture, numPics)
	i := 0
	for i < len(survivors) {
		a := survivors[i]
		b := survivors[rand.Intn(len(survivors))]
		newPics[i] = cross(a, b)
		i++
	}

	for i < len(newPics) {
		a := survivors[rand.Intn(len(survivors))]
		b := survivors[rand.Intn(len(survivors))]
		newPics[i] = cross(a, b)
		i++
	}

	for _, pic := range newPics {
		r := rand.Intn(4)
		for i := 0; i < r; i++ {
			pic.mutate()
		}
	}
	return newPics
}

func (p *picture) mutate() {
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
	//fmt.Println(nodeToMutate)

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
			b := pic.b.Eval(x, y)

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

	picWidth := int(float32(winWidth/cols) * float32(.8))
	picHeight := int(float32(winHeight/rows) * float32(.8))

	pixelsChannel := make(chan pixelResult, numPics)

	evolveButtonText := GetSinglePixelTex(renderer, sdl.Color{255, 255, 255, 0})
	evolveRect := sdl.Rect{int32(float32(winWidth/2) - float32(picWidth)/2), int32(float32(winHeight) - (float32(winHeight) * .1)), int32(picWidth), int32(float32(winHeight) * .08)}
	evolveButton := NewImageButton(renderer, evolveButtonText, evolveRect, sdl.Color{255, 255, 255, 0})

	buttons := make([]*ImageButton, numPics)
	for i := range picTrees {
		go func(i int) {
			pixels := aptToPixels(picTrees[i], picWidth*2, picHeight*2)
			pixelsChannel <- pixelResult{pixels, i}
		}(i)
	}

	// Changd after EP 06 to address MacOSX
	// OSX requires that you consume events for windows to open and work properly
	keyboardState := sdl.GetKeyboardState()
	preKeyboardState := make([]uint8, len(keyboardState))
	for i, v := range keyboardState {
		preKeyboardState[i] = v
	}

	mouseState := GetMouseState()
	guiState := guiState{false, nil}
	for {
		frameStart := time.Now()

		mouseState.Update()
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch e := event.(type) {
			case *sdl.QuitEvent:
				return
			case *sdl.TouchFingerEvent:
				if e.Type == sdl.FINGERDOWN {
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
		if !guiState.zoom {
			select {
			case pixelsAndIndex, ok := <-pixelsChannel:
				if ok {
					tex := pixelsToTexture(renderer, pixelsAndIndex.pixels, picWidth*2, picHeight*2)
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
			for i, button := range buttons {
				if button != nil {
					button.Update(*mouseState)
					if button.WasLeftClicked {
						button.IsSelcted = !button.IsSelcted
					} else if button.WasRightClicked {
						fmt.Println(picTrees[i])
						zoomPixels := aptToPixels(picTrees[i], winWidth*2, winHeight*2)
						zoomTex := pixelsToTexture(renderer, zoomPixels, winWidth*2, winHeight*2)
						guiState.zoomImg = zoomTex
						guiState.zoom = true
					}
					button.Draw(renderer)
				}
			}

			evolveButton.Update(*mouseState)
			if evolveButton.WasLeftClicked {
				selectedPictures := make([]*picture, 0)
				for i, button := range buttons {
					if button.IsSelcted {
						selectedPictures = append(selectedPictures, picTrees[i])
					}
				}
				if len(selectedPictures) != 0 {
					for i := range buttons {
						buttons[i] = nil
					}

					picTrees = evolve(selectedPictures)
					for i := range picTrees {
						go func(i int) {
							pixels := aptToPixels(picTrees[i], picWidth*2, picHeight*2)
							pixelsChannel <- pixelResult{pixels, i}
						}(i)
					}
				}
			}
			evolveButton.Draw(renderer)
		} else {
			if !mouseState.RightButton && mouseState.PrevRightButton {
				guiState.zoom = false
			}
			if keyboardState[sdl.SCANCODE_S] == 0 && preKeyboardState[sdl.SCANCODE_S] != 0 {
				saveTree(state.zoomTree)
			}
			renderer.Copy(guiState.zoomImg, nil, nil)

		}

		renderer.Present()
		for i, v := range keyboardState {
			preKeyboardState[i] = v
		}
		elapsedTime = float32(time.Since(frameStart).Seconds() * 1000)
		if elapsedTime < 5 {
			sdl.Delay(5 - uint32(elapsedTime))
			elapsedTime = float32(time.Since(frameStart).Seconds() * 1000)
		}
	}
}
