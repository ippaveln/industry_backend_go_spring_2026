package main

import (
	"fmt"
	"strconv"
)

func fizzBuzz(n int) (string, error) {
	if n < 0 {
		return "", fmt.Errorf("число должно быть положительным %d", n)
	}
	if n%15 == 0 {
		return "FizzBuzz", nil
	}
	if n%3 == 0 {
		return "Fizz", nil
	}
	if n%5 == 0 {
		return "Buzz", nil
	}
	return strconv.Itoa(n), nil
}
