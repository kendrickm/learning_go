package main

import "fmt"

func sayHello(name string) {
	fmt.Println("Hello ", name)
}

func sayGoodbye(name string) {
	fmt.Println("Goodbye ", name)
}

func beSociable(name string) {
	sayHello(name)
	fmt.Println("How's the weather?")
	sayGoodbye(name)
}

func addOne(x int) int {
	return x + 1
}

func main() {

	beSociable("Bob")
	beSociable("Alice")

	x := 5
	x = addOne(x)
	fmt.Println(x)
}
