package main

import (
	"errors"
	"strconv"
)

func fizzBuzz(n int) (string, error) {
	if n < 0 {
		return "", errors.New("input must be a positive integer")
	}

	switch {
	case n%15 == 0:
		return "FizzBuzz", nil
	case n%3 == 0:
		return "Fizz", nil
	case n%5 == 0:
		return "Buzz", nil
	default:
		return strconv.Itoa(n), nil
	}
}
