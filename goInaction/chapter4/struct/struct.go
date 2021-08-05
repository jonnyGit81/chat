package main

import "fmt"

type Duration int64

func main() {
	var dur Duration
	//dur = int64(1000) error
	dur = Duration(1000) //you need to explicit convert
	fmt.Println(dur)
}
