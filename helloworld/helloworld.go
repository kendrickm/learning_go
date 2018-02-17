package main

import "fmt"

func main() {
	fmt.Println("Hello World")

	for i := 0; i < 10; i++ {
		fmt.Println("The variable i is ", i)
	}

	test := true
	if !test {
		fmt.Println("Hello again")
	} else {
		fmt.Println("Goodbye world")
	}

}
