package main

import "fmt"

func greet(name string) string {
	if name == "" {
		return "Hello, World!"
	} else {
		result := fmt.Sprintf("Hello, %s!", name)
		return result
	}
}
