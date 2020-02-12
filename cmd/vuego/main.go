package main

import (
	"context"
	"log"
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
	vuego.Bind("sum", sum)
	vuego.Bind("timer", timer)

	// vuego.Bind("math.pow", math.Pow)
	// vuego.Bind("math.abs", math.Abs)
	// vuego.Bind("utils.time", map[string]interface{}{"timer": timer})
	// vuego.BindFactory("counter", func(c *vuego.FactoryContext) interface{} {
	// 	go func() {
	// 		<-c.Done
	// 		log.Println("page done")
	// 	}()
	// 	return newCounter()
	// })

	// vuego.Auth = vuego.BasicAuth(func(user, pass string) bool {
	// 	return user == "admin" && pass == "123"
	// })

	err := vuego.ListenAndServe(":8000", http.Dir("./fe/dist"))
	// err := vuego.ListenAndServeTLS(":8000", http.Dir("./fe/dist"), "server.crt", "server.key")
	if err != nil {
		log.Fatal(err)
	}
}
