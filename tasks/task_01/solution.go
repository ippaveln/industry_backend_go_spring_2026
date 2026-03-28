package main

import "fmt"

const (
	defaultName = "World"
)

func greet(name string) string {
	if name == "" {
		name = defaultName
	}
	return fmt.Sprintf("Hello, %s!", name)
}
