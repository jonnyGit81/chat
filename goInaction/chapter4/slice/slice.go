package main

import "fmt"

func main() {
	var s1 []int // nil
	s1 = []int{1, 2, 3, 4}
	fmt.Println(s1)

	var s2 = make([]int, 5) // not nil, 5 len 5 cap, all zero val
	fmt.Println(len(s2))    // 5
	fmt.Println(cap(s2))    // 5
	fmt.Println(s2[0])      // 0

	s3 := []string{99: ""} // 100 len & cap of empty string slice.
	fmt.Println(len(s3))   // 100
	fmt.Println(cap(s3))   // 100

	//nil slice
	var s4 []int           // nil
	fmt.Println(nil == s4) // true
	fmt.Println(len(s4))   // 0

	//empty slice
	s5Empty := make([]int, 0)   // not nil
	fmt.Println(nil == s5Empty) // false

	s6Empty := []int{}
	fmt.Println(nil == s6Empty) // false

	source := []string{"Apple", "Orange", "Plum", "Banana", "Grape"}
	slice := source[2:3:3]
	fmt.Println(slice)
}
