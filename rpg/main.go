package main

import (
	"fmt"

	"github.com/kendrickm/learning_go/rpg/game"
	"github.com/kendrickm/learning_go/rpg/ui2d"
)

func main() {
	ui := &ui2d.UI2d{}
	fmt.Println("Fuck windows")
	game.Run(ui)
}
