package main

import (
	"errors"
	"fmt"
)

func fizzBuzz(n int) (string, error) {
	if n < 0 {
		return "", errors.New("input must be > 0")
	}
	if n%3 == 0 && n%5 == 0 {
		return "FizzBuzz", nil
	}
	if n%3 == 0 {
		return "Fizz", nil
	}
	if n%5 == 0 {
		return "Buzz", nil
	}
	return fmt.Sprintf("%d", n), nil
}
