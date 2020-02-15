package main

import (
	"context"
	"log"
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
			v := write.Call(i)
			if v.Err() != nil {
				log.Printf("timer callback call error: %v", v.Err())
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
	ui := vuego.NewUI(
		vuego.OnlinePort(8000),
		// vuego.OnlineAuth(vuego.BasicAuth(func(user, pass string) bool {
		// 	return user == "admin" && pass == "123"
		// })),
		// vuego.OnlineTLS("server.crt", "server.key"),
	)

	ui.Bind(vuego.Func("sum", sum))
	ui.Bind(vuego.Func("timer", timer))

	// ui.BindFunc("math.pow", math.Pow)
	// ui.BindFunc("math.abs", math.Abs)
	// ui.BindPrefix("utils.time", vuego.Map(map[string]interface{}{"timer": timer}))
	// ui.BindPrefix("counter", vuego.DelayObject(&Counter{}, func(c *vuego.UIContext) vuego.Bindings {
	// 	go func() {
	// 		<-c.Done
	// 		log.Println("page done")
	// 	}()
	// 	return vuego.Object(newCounter())
	// }))


	if err := ui.Run(); err != nil {
		log.Fatal(err)
	}
}