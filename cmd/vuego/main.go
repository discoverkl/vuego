package main

import (
	"log"
	"net/http"
	"github.com/discoverkl/vuego"
)

func add(a, b int) int {
	return a + b
}

func main() {
	http.Handle("/", http.FileServer(http.Dir("./fe/dist")))

	_ = vuego.Bind("add", add)

	if err := http.ListenAndServe(":8000", nil); err != nil {
		log.Fatal(err)
	}

}
