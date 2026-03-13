package main

import "fmt"

func greet(name string) string {
	if len(name) == 0 {
		return "Hello, World!"
	} else {
		return fmt.Sprintf("Hello, %s!", name)
	}
}
