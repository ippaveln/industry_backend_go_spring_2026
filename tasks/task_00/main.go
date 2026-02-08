package main

import "fmt"

func main() {
	res := greet()
	fmt.Println(res)
}

func greet() string {
	return "Hello, World!"
}
