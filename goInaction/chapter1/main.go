package main

import (
	"fmt"
	"sync"
)

var wg sync.WaitGroup

func print(ch <-chan int) {
	for i := range ch {
		fmt.Printf("received channel signal %d\n", i)
		// wg.Done() // wrong place, this is for loop, singaling to wait group will decrement.
	}
	wg.Done()
}

func main() {
	ch := make(chan int)
	// we create 1 go routine and inside print function waiting for receive
	go print(ch)

	wg.Add(1)

	for i := 1; i <= 10; i++ {
		ch <- i
	}

	// Why we need to close here?
	// 1. print function is infinite loop, waiting for received channel data.
	// 2. once we close the for range will be ended.
	// 3. after for loop terminated we tell wait group do decrement.

	// if you forgot to close will hit error, fatal error: all goroutines are asleep - deadlock!
	close(ch)

	// if you not using waitGroup here there is a missing 1 print statement sometimes
	// because the main go routine end first.
	// we maker sure all go routine finish and we resume the main go routine
	wg.Wait()
}
