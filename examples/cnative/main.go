//go:generate go run github.com/markbates/pkger/cmd/pkger -o examples
package main

import (
	"log"

	"github.com/discoverkl/vuego/chrome"
	"github.com/markbates/pkger"
)

//int add(int a, int b);
import "C"

func add(a, b int) int {
	return int(C.add(C.int(a), C.int(b)))
}

func main() {
	win, err := chrome.NewApp(pkger.Dir("/examples/helloworld/fe/dist"), 200, 200, 800, 600)
	// win, err := browser.NewPage(pkger.Dir("/examples/helloworld/fe/dist"))
	if err != nil {
		log.Fatal(err)
	}
	win.Bind("add", add)
	<-win.Done()
}
