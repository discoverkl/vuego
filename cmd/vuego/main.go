package main

import (
	"log"
	"context"
	"net/http"
	"time"

	"github.com/discoverkl/vuego"
)

func add(a, b int) int {
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

func main() {
	vuego.Bind("add", add)
	vuego.Bind("timer", timer)

	if err := vuego.FileServer(http.Dir("./fe/dist"), vuego.Addr(":8000")); err != nil {
		log.Fatal(err)
	}

}
