package main

import (
	"github.com/kendrickm/learning_go/rpg/game"
	"github.com/kendrickm/learning_go/rpg/ui2d"
)

func main() {
	ui := &ui2d.UI2d{}
	game.Run(ui)
}
