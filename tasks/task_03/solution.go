package main

import (
	"fmt"
	"strconv"
)

func fizzBuzz(n int) (string, error) {
	// Типичный fizzBuzz за исключением случаев n<0 -> error != nil

	if n < 0 {
		return "", fmt.Errorf("n<0")
	} else if n%3 == 0 && n%5 == 0 {
		return "FizzBuzz", nil
	} else if n%3 == 0 {
		return "Fizz", nil
	} else if n%5 == 0 {
		return "Buzz", nil
	} else {
		return strconv.Itoa(n), nil
	}
}
