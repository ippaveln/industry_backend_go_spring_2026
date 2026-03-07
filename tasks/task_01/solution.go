package main

import "fmt"

func greet(name string) string {
	if name != "" {
		return fmt.Sprintf("Hello, %s!", name)
	}
	return "Hello, World!"
}
