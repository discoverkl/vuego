//go:generate go run github.com/markbates/pkger/cmd/pkger -o counter
package main

import (
	"log"
	"sync"

	"github.com/discoverkl/vuego/ui"
	"github.com/markbates/pkger"
)

type counter struct {
	sync.Mutex
	count int
}

func (c *counter) Add(n int) {
	c.Lock()
	defer c.Unlock()
	c.count = c.count + n
}

func (c *counter) Value() int {
	c.Lock()
	defer c.Unlock()
	return c.count
}

func main() {
	c := &counter{}

	app := ui.New(
		ui.Root(pkger.Dir("/counter/fe/dist")),
		ui.OnlinePort(8000),
	)

	app.BindFunc("counterAdd", c.Add)
	app.BindFunc("counterValue", c.Value)

	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
