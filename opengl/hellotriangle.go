package main

import (
"github.com/veandco/go-sdl2/sdl"
"github.com/go-gl/gl/v3.3-core/gl"
"fmt"
"github.com/kendrickm/learning_go/gogl"
)

const winWidth = 1080
const winHeight = 720

func main() {
	err := sdl.Init(sdl.INIT_EVERYTHING)
	if err != nil {
		panic(err)
	}
	defer sdl.Quit()
	sdl.GLSetAttribute(sdl.GL_CONTEXT_MAJOR_VERSION, 3)
	sdl.GLSetAttribute(sdl.GL_CONTEXT_MINOR_VERSION, 3)
	sdl.GLSetAttribute(sdl.GL_CONTEXT_PROFILE_MASK, sdl.GL_CONTEXT_PROFILE_CORE)

	window,err := sdl.CreateWindow("Hello Triangle", 200,200, winWidth, winHeight,sdl.WINDOW_OPENGL)
	if err != nil {
		panic(err)
	}

	window.GLCreateContext()
	defer window.Destroy()

	gl.Init()

	fmt.Println("OpenGL Version", gogl.GetVersion())

	shaderProgram, err := gogl.NewShader("shaders/hello.vert", "shaders/hello.frag")
	if err != nil {
		panic (err)
	}

	verticies := []float32{
		0.5,0.5,0.0,     1.0,1.0,
		0.5,-0.5,0.0,    1.0, 0.0,
		-0.5,-0.5,0.0,   0.0,0.0,
		-0.5,0.5,0.0,    0.0,1.0}

    indicies := []uint32 {
    	0,1,3,
    	1,2,3}
	gogl.GenBindBuffer(gl.ARRAY_BUFFER)

	VAO := gogl.GenBindVertexArray()
	gogl.GenBindBuffer(gl.ELEMENT_ARRAY_BUFFER)
	gogl.BufferDataInt(gl.ELEMENT_ARRAY_BUFFER, indicies, gl.STATIC_DRAW)

	gogl.BufferDataFloat(gl.ARRAY_BUFFER, verticies, gl.STATIC_DRAW)


	gl.VertexAttribPointer(0,3,gl.FLOAT,false,5*4,nil)

	gl.EnableVertexAttribArray(0)
	gl.VertexAttribPointer(1,2,gl.FLOAT,false,5*4, gl.PtrOffset(3*4))
	gl.EnableVertexAttribArray(1)
	gogl.UnbindVertexArray()

	var x float32 = 0.0

	for {
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch event.(type) {
			case *sdl.QuitEvent:
				return
			}
		}

		gl.ClearColor(0.0,0.0,0.0,0.0)
		gl.Clear(gl.COLOR_BUFFER_BIT)
		shaderProgram.Use()
		shaderProgram.SetFloat("x",x)
		shaderProgram.SetFloat("y",0)
		gogl.BindVertexArray(VAO)
		gl.DrawElements(gl.TRIANGLES,6,gl.UNSIGNED_INT,gl.PtrOffset(0))
		window.GLSwap()
		shaderProgram.CheckShaderForChanges()

		x = x + .01

	}
}