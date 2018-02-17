package main

import "fmt"

type storyPage struct {
	text     string
	nextPage *storyPage
}

func (page *storyPage) playStory() {
	if page == nil {
		return
	}
	fmt.Println(page.text)
	page.nextPage.playStory()
}

func (page *storyPage) addToEnd(text string) {
	for page.nextPage != nil {
		page = page.nextPage
	}
	page.nextPage = &storyPage{text, nil}
}

func (page *storyPage) addAfter(text string) {
	newPage := &storyPage{text, page.nextPage}
	page.nextPage = newPage
}

func main() {

	page1 := storyPage{"You are standing in an open field west of a white house", nil}
	page1.addToEnd("You climb into the attic. You can't see a thing.")
	page1.addToEnd("You are eaten by a grue")

	page1.addAfter("Testing addAfter")
	page1.playStory()
}
