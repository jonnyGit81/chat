package main

import "fmt"

func main() {
	var m map[int]string // nil
	fmt.Println(m == nil)

	m[1] = "value" //panic: assignment to entry in nil map

	//initialize
	m = map[int]string{10: "value"}
	fmt.Println(m)

	m1 := map[int]string{10: "value"}
	fmt.Println(m1)

	m2 := make(map[string]string) //not nill
	fmt.Println(m2 == nil)        //false
	m2["key"] = "value"           //put
	fmt.Println(m2)

	// Remove the key/value pair for the key "Coral".
	colors := map[string]string{
		"AliceBlue":   "#f0f8ff",
		"Coral":       "#ff7F50",
		"DarkGray":    "#a9a9a9",
		"ForestGreen": "#228b22",
	}
	delete(colors, "Coral")

	//The map key can be a value from any built-in or struct type as long as the value can be used
	// in an expression with the == operator. Slices, functions,
	// and struct types that contain slices can’t be used as map keys. This will produce a compiler error.
}
