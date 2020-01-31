package main

import (
	"context"
	"log"
	"math"
	"net/http"
	"time"

	"github.com/discoverkl/vuego"
)

func sum(a, b int) int {
	return a + b
}

func timer(ctx context.Context, write *vuego.Function) string {
	for i := 0; i < 10; i++ {
		select {
		case <-ctx.Done():
			return "cancel"
		case <-time.After(time.Millisecond * 100):
			err := write.Call(i)
			if err != nil {
				log.Printf("timer callback call error: %v", err)
			}
		}
	}
	return "done"
}

type Counter struct {
	sum int
}

func newCounter() *Counter {
	return &Counter{}
}

func (c *Counter) Add() int {
	c.sum++
	return c.sum
}

func main() {
	vuego.Bind("sum", sum)
	vuego.Bind("timer", timer)
	vuego.Bind("math.pow", math.Pow)
	vuego.Bind("math.abs", math.Abs)

	vuego.BindObject("", newCounter())

	if err := vuego.ListenAndServe(":8000", http.Dir("./fe/dist")); err != nil {
		log.Fatal(err)
	}
}
