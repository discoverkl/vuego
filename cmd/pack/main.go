package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/discoverkl/vuego"
)

func main() {
	flag.Usage = func() {
		fmt.Println("pack [PATH] [OUTPUT_FILE] [PKG_NAME]")
	}
	flag.Parse()

	path := flag.Arg(0)
	output := flag.Arg(1)
	pkg := flag.Arg(2)

	err := vuego.Pack(path, output, pkg)
	if err != nil {
		log.Fatal(err)
	}
}
