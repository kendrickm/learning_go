package ui2d

import (
	"bufio"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"unsafe"

	"github.com/kendrickm/learning_go/rpg/game"
	"github.com/veandco/go-sdl2/img"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
	"github.com/veandco/go-sdl2/mix"
)

func f(p unsafe.Pointer) {}

type sounds struct {
	openingDoors []*mix.Chunk
	footsteps  []*mix.Chunk
}

func playRandomSound(chunks []*mix.Chunk, volume int){
	chunkIndex := rand.Intn(len(chunks))
	fmt.Println("Playing file " + strconv.Itoa(chunkIndex))
	chunks[chunkIndex].Volume(volume)
	chunks[chunkIndex].Play(-1,0)
}

type ui struct {
	sounds sounds
	winWidth  int
	winHeight int

	renderer         *sdl.Renderer
	window           *sdl.Window
	textureAtlas     *sdl.Texture
	textureIndex     map[rune][]sdl.Rect
	preKeyboardState []uint8
	keyboardState    []uint8
	r                *rand.Rand
	centerX          int
	centerY          int

	levelChan chan *game.Level
	inputChan chan *game.Input

	fontMedium *ttf.Font
	fontLarge  *ttf.Font
	fontSmall  *ttf.Font

	eventBackground 		  *sdl.Texture
	groundInventoryBackground *sdl.Texture

	string2TexSmall map[string]*sdl.Texture
	string2TexMed   map[string]*sdl.Texture
	string2TexLarge map[string]*sdl.Texture
}

func NewUI(inputChan chan *game.Input, levelChan chan *game.Level) *ui {
	ui := &ui{}
	ui.inputChan = inputChan
	ui.levelChan = levelChan
	ui.string2TexSmall = make(map[string]*sdl.Texture)
	ui.string2TexMed = make(map[string]*sdl.Texture)
	ui.string2TexLarge = make(map[string]*sdl.Texture)
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

	ui.fontSmall, err = ttf.OpenFont("ui2d/assets/gothic.ttf", int(float64(ui.winWidth)*.015))
	if err != nil {
		panic(err)
	}

	ui.fontMedium, err = ttf.OpenFont("ui2d/assets/gothic.ttf", 32)
	if err != nil {
		panic(err)
	}

	ui.fontLarge, err = ttf.OpenFont("ui2d/assets/gothic.ttf", 64)
	if err != nil {
		panic(err)
	}

	ui.eventBackground = ui.GetSinglePixelTex(sdl.Color{0, 0, 0, 128})
	ui.eventBackground.SetBlendMode(sdl.BLENDMODE_BLEND)

	ui.groundInventoryBackground = ui.GetSinglePixelTex(sdl.Color{255, 0, 0, 128})
	ui.groundInventoryBackground.SetBlendMode(sdl.BLENDMODE_BLEND)

	err = mix.OpenAudio(22050, mix.DEFAULT_FORMAT,2,4096)
	if err != nil {
		panic(err)
	}

	mus,err :=  mix.LoadMUS("ui2d/assets/ambient.ogg")
	if err != nil {
		panic(err)
	}
	mus.Play(-1)

	footstepBase := "ui2d/assets/footstep0"
	for i := 0; i<10; i++ {
		footstepFile := footstepBase + strconv.Itoa(i) + ".ogg"
		fmt.Println("Loading " + footstepFile)
		footstepSound, err := mix.LoadWAV(footstepFile)
		if err != nil {
		panic(err)
		}
		ui.sounds.footsteps = append(ui.sounds.footsteps, footstepSound)
	}

	doorOpen1, err := mix.LoadWAV("ui2d/assets/doorOpen_1.ogg")
	if err != nil {
		panic(err)
	}
	ui.sounds.openingDoors = append(ui.sounds.openingDoors, doorOpen1)

	doorOpen2, err := mix.LoadWAV("ui2d/assets/doorOpen_2.ogg")
	if err != nil {
		panic(err)
	}
	ui.sounds.openingDoors = append(ui.sounds.openingDoors, doorOpen2)

	return ui
}

type FontSize int

const (
	FontSmall FontSize = iota
	FontMedium
	FontLarge
)

func (ui *ui) stringToTexture(s string, size FontSize, color sdl.Color) *sdl.Texture {

	var font *ttf.Font
	switch size {
	case FontSmall:
		font = ui.fontSmall
		tex, exists := ui.string2TexSmall[s]
		if exists {
			return tex
		}
	case FontMedium:
		font = ui.fontMedium
		tex, exists := ui.string2TexMed[s]
		if exists {
			return tex
		}
	case FontLarge:
		font = ui.fontLarge
		tex, exists := ui.string2TexLarge[s]
		if exists {
			return tex
		}
	}
	fontSurface, err := font.RenderUTF8Blended(s, sdl.Color{255, 0, 0, 0})
	if err != nil {
		panic(err)
	}
	fontTexture, err := ui.renderer.CreateTextureFromSurface(fontSurface)
	if err != nil {
		panic(err)
	}

	switch size {
	case FontSmall:
		ui.string2TexSmall[s] = fontTexture
	case FontMedium:
		ui.string2TexMed[s] = fontTexture
	case FontLarge:
		ui.string2TexLarge[s] = fontTexture
	}

	return fontTexture
}

func (ui *ui) loadTextureIndex() {
	ui.textureIndex = make(map[rune][]sdl.Rect)
	file, err := os.Open("ui2d/assets/asset-index.txt")
	if err != nil {
		panic(err)
	}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)
		tileRune := rune(line[0])
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

		ui.textureIndex[tileRune] = rects
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
		panic(err)
	}
	err = ttf.Init()
	if err != nil {
		panic(err)
	}

	err = mix.Init(mix.INIT_OGG)
		if err != nil {
		panic(err)
	}

}

func (ui *ui) keyDownOnce(key uint8) bool {
	return ui.keyboardState[key] == 1 && ui.preKeyboardState[key] == 0
}

// Check for key pressed and then released
func (ui *ui) keyPressed(key uint8) bool {
	return ui.keyboardState[key] == 0 && ui.preKeyboardState[key] == 1
}

func (ui *ui) GetSinglePixelTex(color sdl.Color) *sdl.Texture {
	tex, err := ui.renderer.CreateTexture(sdl.PIXELFORMAT_ABGR8888, sdl.TEXTUREACCESS_STATIC, 1, 1)
	if err != nil {
		panic(err)
	}
	pixels := make([]byte, 4)
	pixels[0] = color.R
	pixels[1] = color.G
	pixels[2] = color.B
	pixels[3] = color.A

	tex.Update(nil, unsafe.Pointer(&pixels[3]), 4)
	err = tex.SetBlendMode(sdl.BLENDMODE_BLEND)
	if err != nil {
		panic(err)
	}
	return tex
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
				switch newLevel.LastEvent {
				case game.Move:
					fmt.Println("Playing footsteps")
					playRandomSound(ui.sounds.footsteps, 10)
				case game.DoorOpen:
					fmt.Println("Playing opening door")
					playRandomSound(ui.sounds.openingDoors, 32)
				default:
				}
				ui.Draw(newLevel)
			}
		default:
		}

		if sdl.GetKeyboardFocus() == ui.window || sdl.GetMouseFocus() == ui.window {
			var input game.Input
			if ui.keyDownOnce(sdl.SCANCODE_UP) {
				input.Typ = game.Up
			} else if ui.keyDownOnce(sdl.SCANCODE_DOWN) {
				input.Typ = game.Down
			} else if ui.keyDownOnce(sdl.SCANCODE_RIGHT) {
				input.Typ = game.Right
			} else if ui.keyDownOnce(sdl.SCANCODE_LEFT) {
				input.Typ = game.Left
			} else if ui.keyDownOnce(sdl.SCANCODE_T) {
				input.Typ = game.TakeAll
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
		diff := level.Player.X - (ui.centerX+limit)
		ui.centerX += diff
	} else if level.Player.X < ui.centerX-limit {
		diff := (ui.centerX-limit) - level.Player.X
		ui.centerX -= diff
	} else if level.Player.Y > ui.centerY+limit {
		diff :=  level.Player.Y - (ui.centerY+limit)
		ui.centerY += diff
	} else if level.Player.Y < ui.centerY-limit {
		diff := (ui.centerY-limit) - level.Player.Y
		ui.centerY -= diff
	}

	offsetX := int32((ui.winWidth / 2) - ui.centerX*32)
	offsetY := int32((ui.winHeight / 2) - ui.centerY*32)
	ui.renderer.Clear()
	ui.r.Seed(2)
	for y, row := range level.Map {
		for x, tile := range row {
			if tile.Rune != game.Blank {
				srcRects := ui.textureIndex[tile.Rune]
				srcRect := srcRects[ui.r.Intn(len(srcRects))]
				if tile.Visible || tile.Seen {
					dstRect := sdl.Rect{int32(x*32) + offsetX, int32(y*32) + offsetY, 32, 32}
					pos := game.Pos{x, y}
					if level.Debug[pos] {
						ui.textureAtlas.SetColorMod(128, 0, 0)
					} else if tile.Seen && !tile.Visible {
						ui.textureAtlas.SetColorMod(128, 128, 128)
					} else {
						ui.textureAtlas.SetColorMod(255, 255, 255)
					}
					ui.renderer.Copy(ui.textureAtlas, &srcRect, &dstRect)

					if tile.OverlayRune != game.Blank {
						// Todo what if there are multiple varients for overlay images?
						srcRect := ui.textureIndex[tile.OverlayRune][0]
						ui.renderer.Copy(ui.textureAtlas, &srcRect, &dstRect)
					}
				}
			}
		}
	}
	ui.textureAtlas.SetColorMod(255, 255, 255)

	//Render Monsters
	for pos, monster := range level.Monsters {
		if level.Map[pos.Y][pos.X].Visible {
			monsterSrcRect := ui.textureIndex[monster.Rune][0]
			ui.renderer.Copy(ui.textureAtlas, &monsterSrcRect, &sdl.Rect{X: int32(pos.X)*32 + offsetX, Y: int32(pos.Y)*32 + offsetY, W: 32, H: 32})
		}
	}
	//Render Items
	for pos, items := range level.Items {
		if level.Map[pos.Y][pos.X].Visible {
			for _, item := range items {
				itemSrcRect := ui.textureIndex[item.Rune][0]
				ui.renderer.Copy(ui.textureAtlas, &itemSrcRect, &sdl.Rect{X: int32(pos.X)*32 + offsetX, Y: int32(pos.Y)*32 + offsetY, W: 32, H: 32})
		    }
		}
	}
	playerSrcRect := ui.textureIndex['@'][0]
	ui.renderer.Copy(ui.textureAtlas, &playerSrcRect, &sdl.Rect{X: int32(level.Player.X)*32 + offsetX, Y: int32(level.Player.Y)*32 + offsetY, W: 32, H: 32})

	// TODO Scroll from bottom up
	// TODO add border/background
	textStart := int32(float64(ui.winHeight) * 0.68)
	textWidth := int32(float64(ui.winWidth) * 0.25)
	ui.renderer.Copy(ui.eventBackground, nil, &sdl.Rect{0, textStart, textWidth, int32(ui.winHeight) - textStart})
	i := level.EventPos
	count := 0
	_, fontSizeY, _ := ui.fontSmall.SizeUTF8("A")
	for {
		event := level.Events[i]
		if event != "" {
			tex := ui.stringToTexture(event, FontSmall, sdl.Color{255, 0, 0, 0})
			_, _, w, h, _ := tex.Query()
			ui.renderer.Copy(tex, nil, &sdl.Rect{5, int32(count*fontSizeY) + textStart, w, h})
		}
		i = (i + 1) % (len(level.Events))
		count++
		if i == level.EventPos {
			break
		}
	}// Event UI end

	//Inventory UI
	groundInvStart := int32(float64(ui.winWidth) * 0.9)
	groundInvWidth := int32(ui.winWidth) - groundInvStart
	ui.renderer.Copy(ui.groundInventoryBackground, nil, &sdl.Rect{groundInvStart, int32(ui.winWidth - 32), groundInvWidth, int32(32)})
	items := level.Items[level.Player.Pos]
	for i, item := range items {
		itemSrcRect := ui.textureIndex[item.Rune][0]
		ui.renderer.Copy(ui.textureAtlas, &itemSrcRect, &sdl.Rect{X:int32(ui.winWidth-32 -i*32), Y: int32(ui.winHeight-32), W: 32, H: 32})  
	}

	ui.renderer.Present()

}
