package main

import "fmt"

func main() {
	var arr [3]int
	arr[0] = 1

	arrs := [5]int{10, 20, 30, 40, 50}
	fmt.Println(arrs)

	// Declare an integer array.
	// Initialize each element with a specific value.
	// Capacity is determined based on the number of values initialized.
	array := [...]int{10, 20, 30, 40, 50}
	fmt.Println(len(array))

	// Declare an integer array of five elements.
	// Initialize index 1 and 2 with specific values.
	// The rest of the elements contain their zero value.
	array2 := [5]int{1: 10, 2: 20}
	fmt.Println(array2)

	// Declare an integer pointer array of five elements.
	// Initialize index 0 and 1 of the array with integer pointers.
	array3 := [5]*int{0: new(int), 1: new(int)}

	// Assign values to index 0 and 1.
	*array3[0] = 10
	*array3[1] = 20
}
