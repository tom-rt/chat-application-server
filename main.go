package main

import (
	"flag"
	"fmt"
)

func main() {
	var port string
	flag.StringVar(&port, "port", "8080", "wrong port value")

	fmt.Println("Running on port", port)

}
