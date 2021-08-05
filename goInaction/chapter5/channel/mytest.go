package main

import (
	"fmt"
	"sync"
)

func main() {
	ch := make(chan string)

	var wg sync.WaitGroup
	wg.Add(2)

	go goRoutineProcessWorking("1", ch, &wg)

	go goRoutineProcessWorking("2", ch, &wg)

	ch <- "BOOM"
	//ch <- "BOOM"

	wg.Wait()
}

func goRoutineProcessWorking(t string, ch chan string, wg *sync.WaitGroup) {
	defer wg.Done()
	fmt.Printf("Thread %s will starting in 10 seconds\n", t)
	//time.Sleep(time.Second * 10)
	fmt.Printf("Thread %s receiving paket %v\n", t, <-ch)
}
