package main

import (
	"fmt"
	"image/png"
	"math"
	"math/rand"
	"os"
	"sort"
	"time"

	"github.com/kendrickm/noise"
	. "github.com/kendrickm/vec3"
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

type balloon struct {
	tex  *sdl.Texture
	pos  Vector3
	dir  Vector3
	w, h int

	exploding         bool
	exploded          bool
	explosionStart    time.Time
	explosionInterval float32
	explosionTexture  *sdl.Texture
}

func newBalloon(tex *sdl.Texture, pos, dir Vector3, explosionTexture *sdl.Texture) *balloon {
	_, _, w, h, err := tex.Query()
	if err != nil {
		panic(err)
	}
	return &balloon{tex, pos, dir, int(w), int(h), false, false, time.Now(), 20, explosionTexture}
}

type balloonArray []*balloon

func (balloons balloonArray) Len() int {
	return len(balloons)
}

func (balloons balloonArray) Swap(i, j int) {
	balloons[i], balloons[j] = balloons[j], balloons[i]
}

func (balloons balloonArray) Less(i, j int) bool {
	diff := balloons[i].pos.Z - balloons[j].pos.Z
	return diff < -0.5
}

func (balloon *balloon) getScale() float32 {
	return (balloon.pos.Z/200 + 1) / 2
}

func (balloon *balloon) getCircle() (x, y, r float32) {
	x = balloon.pos.X
	y = balloon.pos.Y - 30*balloon.getScale()
	r = float32(balloon.w) / 2 * balloon.getScale()

	return x, y, r
}

func (balloon *balloon) update(elapsedTime float32,
	currentMouseState mouseState,
	prevMouseState mouseState,
	audioState *audioState) {
	numAnimations := 16
	animationsElapsed := float32(time.Since(balloon.explosionStart).Seconds() * 1000)
	animationIndex := numAnimations - 1 - int(animationsElapsed/balloon.explosionInterval)
	if animationIndex < 0 {
		balloon.exploding = false
		balloon.exploded = true
	}

	if !prevMouseState.leftButton && currentMouseState.leftButton {
		x, y, r := balloon.getCircle()
		mouseX := currentMouseState.x
		mouseY := currentMouseState.y

		xDiff := float32(mouseX) - x
		yDiff := float32(mouseY) - y
		dist := float32(math.Sqrt(float64(xDiff*xDiff + yDiff*yDiff)))
		if dist < r {
			sdl.ClearQueuedAudio(audioState.deviceID)
			sdl.QueueAudio(audioState.deviceID, audioState.explosionBytes)
			sdl.PauseAudioDevice(audioState.deviceID, false)
			balloon.exploding = true
			balloon.explosionStart = time.Now()
		}
	}
	p := Add(balloon.pos, Mult(balloon.dir, elapsedTime))

	if p.X < 0 || p.X > float32(winWidth) {
		balloon.dir.X = -balloon.dir.X
	}

	if p.Y < 0 || p.Y > float32(winHeight) {
		balloon.dir.Y = -balloon.dir.Y
	}

	if p.Z < 0 || p.Z > float32(winDepth) {
		balloon.dir.Z = -balloon.dir.Z
	}

	balloon.pos = Add(balloon.pos, Mult(balloon.dir, elapsedTime))
}

func (balloon *balloon) draw(renderer *sdl.Renderer) {
	scale := balloon.getScale()
	newW := int32(float32(balloon.w) * scale)
	newH := int32(float32(balloon.h) * scale)
	x := int32(balloon.pos.X - float32(newW)/2)
	y := int32(balloon.pos.Y - float32(newH)/2)
	rect := &sdl.Rect{x, y, newW, newH}
	renderer.Copy(balloon.tex, nil, rect)

	if balloon.exploding {
		numAnimations := 16
		animationsElapsed := float32(time.Since(balloon.explosionStart).Seconds() * 1000)
		animationIndex := numAnimations - 1 - int(animationsElapsed/balloon.explosionInterval)
		animationX := animationIndex % 4
		animationY := 64 * ((animationIndex - animationX) / 4)
		animationX *= 64
		animationRect := &sdl.Rect{int32(animationX), int32(animationY), 64, 64}
		rect.X = rect.W / 2
		rect.Y = rect.H / 2
		rect.H *= 2
		rect.W *= 2
		renderer.Copy(balloon.explosionTexture, animationRect, rect)
	}
}

type rgba struct {
	r, g, b byte
}

func lerp(b1 byte, b2 byte, pct float32) byte {
	return byte(float32(b1) + pct*(float32(b2)-float32(b1)))
}

func colorLerp(c1, c2 rgba, pct float32) rgba {
	return rgba{lerp(c1.r, c2.r, pct), lerp(c1.g, c2.g, pct), lerp(c1.b, c2.b, pct)}
}

func getGradient(c1, c2 rgba) []rgba {
	result := make([]rgba, 256)
	for i := range result {
		pct := float32(i) / float32(255)
		result[i] = colorLerp(c1, c2, pct)
	}

	return result
}

func getDuelGradient(c1, c2, c3, c4 rgba) []rgba {
	result := make([]rgba, 256)
	for i := range result {
		pct := float32(i) / float32(255)
		if pct < 0.5 {
			result[i] = colorLerp(c1, c2, pct*float32(2))
		} else {
			result[i] = colorLerp(c3, c4, pct*float32(1.5)-float32(.5))
		}

	}

	return result
}

func clamp(min, max, v int) int {
	if v < min {
		v = min
	} else if v > max {
		v = max
	}
	return v
}

func rescaleAndDraw(noise []float32, min, max float32, gradient []rgba, w, h int) []byte {
	result := make([]byte, w*h*4)

	scale := 255.0 / (max - min)
	offset := min * scale

	for i := range noise {
		noise[i] = noise[i]*scale - offset
		// b := byte(noise[i])
		c := gradient[clamp(0, 255, int(noise[i]))]
		p := i * 4
		result[p] = c.r
		result[p+1] = c.g
		result[p+2] = c.b
	}
	return result
}

func pixelsToTexture(renderer *sdl.Renderer, pixels []byte, w, h int) *sdl.Texture {
	tex, err := renderer.CreateTexture(sdl.PIXELFORMAT_ABGR8888, sdl.TEXTUREACCESS_STREAMING, int32(w), int32(h))
	if err != nil {
		panic(err)
	}
	tex.Update(nil, pixels, w*4)
	return tex
}

func imgFileToTexture(renderer *sdl.Renderer, filename string) *sdl.Texture {

	infile, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer infile.Close()

	img, err := png.Decode(infile)
	if err != nil {
		panic(err)
	}

	w := img.Bounds().Max.X
	h := img.Bounds().Max.Y

	pixels := make([]byte, w*h*4)
	bIndex := 0

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			r, g, b, a := img.At(x, y).RGBA()
			pixels[bIndex] = byte(r / 256)
			bIndex++
			pixels[bIndex] = byte(g / 256)
			bIndex++
			pixels[bIndex] = byte(b / 256)
			bIndex++
			pixels[bIndex] = byte(a / 256)
			bIndex++
		}
	}
	tex := pixelsToTexture(renderer, pixels, w, h)
	err = tex.SetBlendMode(sdl.BLENDMODE_BLEND)
	if err != nil {
		panic(err)
	}

	return tex
}

func loadBaloons(renderer *sdl.Renderer, numBalloons int) []*balloon {

	explosionTexture := imgFileToTexture(renderer, "explosion.png")

	balloonStr := []string{"balloon_red.png", "balloon_green.png", "balloon_blue.png"}
	balloonsTextures := make([]*sdl.Texture, len(balloonStr))

	for i, bstr := range balloonStr {
		balloonsTextures[i] = imgFileToTexture(renderer, bstr)
	}

	balloons := make([]*balloon, numBalloons)
	for i := range balloons {
		tex := balloonsTextures[i%3]
		pos := Vector3{rand.Float32() * float32(winWidth), rand.Float32() * float32(winHeight), rand.Float32() * float32(winDepth)}
		dir := Vector3{rand.Float32()*.5 - 0.25, rand.Float32()*.5 - 0.25, rand.Float32()*.25 - .25/2}
		balloons[i] = newBalloon(tex, pos, dir, explosionTexture)
	}

	return balloons
}

func main() {

	sdl.LogSetAllPriority(sdl.LOG_PRIORITY_VERBOSE)
	err := sdl.Init(sdl.INIT_EVERYTHING)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer sdl.Quit()

	window, err := sdl.CreateWindow("Exploding Balloons", 200, 200,
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

	explosionBytes, audioSpec := sdl.LoadWAV("explode.wav")
	audioID, err := sdl.OpenAudioDevice("", false, audioSpec, nil, 0)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer sdl.FreeWAV(explosionBytes)

	audioState := audioState{explosionBytes, audioID, audioSpec}

	noise, min, max := noise.MakeNoise(noise.FBM, .009, 3, .5, 3, winWidth, winHeight)
	cloudGradient := getGradient(rgba{0, 0, 255}, rgba{255, 255, 255})
	cloudPixels := rescaleAndDraw(noise, min, max, cloudGradient, winWidth, winHeight)
	cloudTexture := pixelsToTexture(renderer, cloudPixels, winWidth, winHeight)
	balloons := loadBaloons(renderer, 25)
	var elapsedTime float32
	currentMouseState := getMouseState()
	prevMouseState := currentMouseState

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

		renderer.Copy(cloudTexture, nil, nil)
		for _, balloon := range balloons {
			balloon.update(elapsedTime, currentMouseState, prevMouseState, &audioState)
		}
		sort.Stable(balloonArray(balloons))
		for _, balloon := range balloons {
			balloon.draw(renderer)
		}

		renderer.Present()
		elapsedTime = float32(time.Since(frameStart).Seconds() * 1000)
		// fmt.Println("ms per frame: ", elapsedTime)
		if elapsedTime < 5 {
			sdl.Delay(5 - uint32(elapsedTime))
			elapsedTime = float32(time.Since(frameStart).Seconds() * 1000)
		}

		prevMouseState = currentMouseState
	}
}
