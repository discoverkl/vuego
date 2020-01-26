//go:generate go run github.com/markbates/pkger/cmd/pkger -o examples
package main

import (
	"log"

	"github.com/discoverkl/vuego"
	"github.com/discoverkl/vuego/browser"
	"github.com/discoverkl/vuego/chrome"
	"github.com/markbates/pkger"
)

func add(a, b int) int {
	return a + b
}

func main() {
	runWebServer()
	// runLocalPage()
	// runNativeApp()
}

// run a normal web server
func runWebServer() {
	vuego.Bind("add", add)

	addr := ":8000"
	log.Printf("listen on: %s", addr)
	if err := vuego.ListenAndServe(addr, pkger.Dir("/examples/helloworld/fe/dist")); err != nil {
		log.Fatal(err)
	}
}

// run a local web server in background and open its' serving url with your default web browser
func runLocalPage() {
	win, err := browser.NewPage(pkger.Dir("/examples/helloworld/fe/dist"))
	if err != nil {
		log.Fatal(err)
	}
	win.Bind("add", add)
	<-win.Done()
}

// run a local web server in background and open its' serving url within a native app (which is a chrome process)
func runNativeApp() {
	win, err := chrome.NewApp(pkger.Dir("/examples/helloworld/fe/dist"), 200, 200, 800, 600)
	if err != nil {
		log.Fatal(err)
	}
	win.Bind("add", add)
	<-win.Done()
}
