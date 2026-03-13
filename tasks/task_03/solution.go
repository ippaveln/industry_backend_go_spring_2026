package main

import "fmt"

func fizzBuzz(n int) (string, error) {
	if n < 0 {
		return "", fmt.Errorf("negative number: %d", n)
	} else if n%3 == 0 && n%5 == 0 {
		return "FizzBuzz", nil
	} else if n%3 == 0 {
		return "Fizz", nil
	} else if n%5 == 0 {
		return "Buzz", nil
	} else {
		return fmt.Sprint(n), nil
	}
}
