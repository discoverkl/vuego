//go:generate go run github.com/markbates/pkger/cmd/pkger -o examples/native
package main

import (
	"log"

	"github.com/discoverkl/vuego/browser"
	"github.com/markbates/pkger"
)

func add(a, b int) int {
	return a + b
}

func main() {
	// win, err := chrome.NewApp(pkger.Dir("/examples/native/fe/dist"), 200, 200, 800, 600)
	win, err := browser.NewPage(pkger.Dir("/examples/native/fe/dist"))
	if err != nil {
		log.Fatal(err)
	}

	win.Bind("add", add)

	// <-time.After(2 * time.Second)
	// win.Close()
	<-win.Done()
}
