package main

import (
	"runtime"

	"github.com/kendrickm/learning_go/rpg/game"
	"github.com/kendrickm/learning_go/rpg/ui2d"
)

func main() {
	game := game.NewGame(3, "game/maps/level1.map")

	for i := 0; i < 3; i++ {
		go func(i int) {
			runtime.LockOSThread()
			ui := ui2d.NewUI(game.InputChan, game.LevelChans[i])
			ui.Run()
		}(i)
	}
	game.Run()

	// go func() { game.Run() }()
	// ui := ui2d.NewUI(game.InputChan, game.LevelChans[0])
	// ui.Run()
}
