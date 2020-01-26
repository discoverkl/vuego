//go:generate go run github.com/markbates/pkger/cmd/pkger -o examples
package main

import (
	"log"
	"sync"

	"github.com/discoverkl/vuego"
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
	vuego.Bind("counterAdd", c.Add)
	vuego.Bind("counterValue", c.Value)

	addr := ":8000"
	log.Printf("listen on: %s", addr)
	if err := vuego.ListenAndServe(addr, pkger.Dir("/examples/counter/fe/dist")); err != nil {
		log.Fatal(err)
	}
}
