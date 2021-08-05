package main

import "fmt"

type employee struct {
	name string
}

func main() {
	emp1 := employee{"Atan"}
	emp2 := employee{"Atan"}
	fmt.Println(emp1 == emp2)
}
