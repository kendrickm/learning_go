package main

import (
	"github.com/kendrickm/learning_go/rpg/game"
	"github.com/kendrickm/learning_go/rpg/ui2d"
)

func main() {
	// TODO When we need multiple UI Support refactor event polling to it's own component
	// and run only on main thread
	game := game.NewGame(1, "game/maps/level1.map")
	go func() {
		game.Run()

	}()
	ui := ui2d.NewUI(game.InputChan, game.LevelChans[0])
	ui.Run()

	// go func() { game.Run() }()
	// ui := ui2d.NewUI(game.InputChan, game.LevelChans[0])
	// ui.Run()
}
