package main

import (
	"bufio"
	"fmt"
	"os"
)

type storyNode struct {
	text    string
	yesPath *storyNode
	noPath  *storyNode
}

func (node *storyNode) printStory(depth int) {
	for i := 0; i < depth; i++ {
		fmt.Print("  ")
	}

	fmt.Print(node.text)
	fmt.Println()

	if node.yesPath != nil {
		node.yesPath.printStory(depth + 1)
	}
	if node.noPath != nil {
		node.noPath.printStory(depth + 1)
	}
}

func (node *storyNode) play() {
	fmt.Println(node.text)
	scanner := bufio.NewScanner(os.Stdin)

	if node.yesPath != nil && node.noPath != nil {
		for {
			scanner.Scan()
			answer := scanner.Text()

			if answer == "yes" {
				node.yesPath.play()
				break
			} else if answer == "no" {
				node.noPath.play()
				break
			} else {
				fmt.Println("That answer was not an option! Please use only yes or no.")
			}
		}
	}
}

func main() {

	root := storyNode{"You are at the enterance to a dark cave. Do you want to enter?", nil, nil}
	winning := storyNode{"You have won!", nil, nil}
	losing := storyNode{"You have lost!", nil, nil}
	root.yesPath = &losing
	root.noPath = &winning

	// root.play()
	root.printStory(0)
}
