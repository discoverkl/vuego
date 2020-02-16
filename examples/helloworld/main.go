//go:generate go run github.com/markbates/pkger/cmd/pkger
package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/discoverkl/vuego/ui"
	"github.com/discoverkl/vuego/browser"
	"github.com/discoverkl/vuego/chrome"
	"github.com/markbates/pkger"
)

const promptText = `
*** Commands ***

1: WebServer - run a normal web server
2: LocalPage - run a local web server in background and open its' serving url with your default web browser
3: NativeApp - run a local web server in background and open its' serving url within a native app (which is a chrome process)

Please enter (1-3)? `

func add(a, b int) int {
	return a + b
}

func main() {
	for {
		fmt.Print(promptText)
		ch, _, err := bufio.NewReader(os.Stdin).ReadRune()
		if err != nil {
			if errors.Is(err, io.EOF) {
				fmt.Println()
				os.Exit(0)
			}
			log.Fatal(err)
		}

		switch ch {
		case '1':
			runWebServer()
		case '2':
			runLocalPage()
		case '3':
			runNativeApp()
		case 'q':
		default:
			continue
		}
		return
	}
}

// site root (http.FileSystem)
const www = pkger.Dir("/fe/dist")

func runWebServer() {
	ui.Bind("add", add)

	addr := ":8000"
	log.Printf("listen on: %s", addr)
	if err := ui.ListenAndServe(addr, www); err != nil {
		log.Fatal(err)
	}
}

func runLocalPage() {
	win, err := browser.NewPage(www)
	if err != nil {
		log.Fatal(err)
	}
	win.Bind("add", add)
	<-win.Done()
}

func runNativeApp() {
	win, err := chrome.NewApp(pkger.Dir(www), 200, 200, 800, 600)
	if err != nil {
		log.Fatal(err)
	}
	win.Bind("add", add)
	<-win.Done()
}
