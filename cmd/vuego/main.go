package main

import (
	"log"
	"net/http"
	"time"

	"github.com/discoverkl/vuego"
)

func add(a, b int) int {
	return a + b
}

func timer(write *vuego.Function) {
	for i := 0; i < 3; i++ {
		<-time.After(time.Millisecond * 10)
		write.Call(i)
	}
}

func main() {
	http.Handle("/", http.FileServer(http.Dir("./fe/dist")))

	_ = vuego.Bind("add", add)
	vuego.Bind("timer", timer)

	if err := http.ListenAndServe(":8000", nil); err != nil {
		log.Fatal(err)
	}

}
