module github.com/kendrickm/learning_go/rpg

replace github.com/kendrickm/learning_go/rpg/game => ./game

replace github.com/kendrickm/learning_go/rpg/ui2d => ./ui2d

go 1.19

require (
	github.com/kendrickm/learning_go/rpg/game v0.0.0-00010101000000-000000000000
	github.com/kendrickm/learning_go/rpg/ui2d v0.0.0-00010101000000-000000000000
)
