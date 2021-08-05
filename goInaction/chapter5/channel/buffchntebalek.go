package main

import "fmt"

func main() {
	h := make(chan string)

	go received(h)
	h <- "Data"

	//for {}
}

func received(ch chan string) {
	fmt.Println(<-ch)
}
