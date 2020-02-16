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
	ui.Bind("counterAdd", c.Add)
	ui.Bind("counterValue", c.Value)

	addr := ":8000"
	log.Printf("listen on: %s", addr)
	if err := ui.ListenAndServe(addr, pkger.Dir("/counter/fe/dist")); err != nil {
		log.Fatal(err)
	}
}
