package main

import (
	"log"
	"context"
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

func main() {
	vuego.Bind("sum", sum)
	vuego.Bind("timer", timer)

	if err := vuego.ListenAndServe(":8000", http.Dir("./fe/dist")); err != nil {
		log.Fatal(err)
	}

}
