package main

import (
	"log"
	"net/http"

	"github.com/discoverkl/vuego/chrome"
	"github.com/discoverkl/vuego"
)

func add(a, b int) int {
	return a + b
}

func main() {
	vuego.Bind("add", add)

	win, err := chrome.New(http.Dir("./fe/dist"), "", 200, 200, 800, 600)
	if err != nil {
		log.Fatal(err)
	}

	// <-time.After(3 * time.Second)
	// win.Close()
	<-win.Done()
}
