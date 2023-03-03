module github.com/kendrickm/learning_go/evolvingpictures

require (
	github.com/kendrickm/learning_go/evolvingpictures/apt v1.2.3
	github.com/kendrickm/learning_go/evolvingpictures/gui v1.2.3
	github.com/kendrickm/learning_go/noise v1.2.3
	github.com/veandco/go-sdl2 v0.4.24
)

replace github.com/kendrickm/learning_go/evolvingpictures/apt => ./apt

replace github.com/kendrickm/learning_go/evolvingpictures/gui => ./gui

replace github.com/kendrickm/learning_go/noise => ../noise

go 1.17
