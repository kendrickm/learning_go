package ui2d

import (
	"fmt"

	"github.com/kendrickm/learning_go/rpg/game"
)

type UI2d struct {
}

func (*UI2d) Draw(level *game.Level) {
	fmt.Println("We did something")
}
