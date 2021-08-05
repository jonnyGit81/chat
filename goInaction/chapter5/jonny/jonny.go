package main

import "fmt"

func main() {
	h := make(chan string, 1)

	h <- "Data"
	go received(h)

	//for {}
}

func received(ch chan string) {
	fmt.Println(<-ch)
}
