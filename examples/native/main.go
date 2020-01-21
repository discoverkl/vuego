package main

import (
	"log"
	"net/http"

	"github.com/discoverkl/vuego/chrome"
)

func add(a, b int) int {
	return a + b
}

func main() {
	win, err := chrome.NewApp(http.Dir("./fe/dist"), 200, 200, 800, 600)
	// win, err := chrome.NewPage(http.Dir("./fe/dist"))
	// win, err := browser.NewPage(http.Dir("./fe/dist"))
	if err != nil {
		log.Fatal(err)
	}

	win.Bind("add", add)

	// <-time.After(2 * time.Second)
	// win.Close()
	<-win.Done()
}
