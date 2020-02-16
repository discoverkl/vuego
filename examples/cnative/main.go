//go:generate go run github.com/markbates/pkger/cmd/pkger -o cnative
package main

import (
	"log"

	"github.com/discoverkl/vuego/ui"
	"github.com/markbates/pkger"
)

//int add(int a, int b);
import "C"

func add(a, b int) int {
	return int(C.add(C.int(a), C.int(b)))
}

func main() {
	app := ui.New(
		ui.Root(pkger.Dir("/helloworld/fe/dist")),
		ui.OnlinePort(8000),
	)
	app.BindFunc("add", add)
	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
