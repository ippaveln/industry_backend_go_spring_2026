package main

func greet(name string) string {
	if len(name) == 0 {
		return "Hello, World!"
	}

	return "Hello, " + name + "!"
}
