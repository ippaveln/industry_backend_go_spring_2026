package main

import (
	"errors"
	"strconv"
)

var ErrNegativeNumber = errors.New("number must be greater than 0")

func fizzBuzz(n int) (string, error) {
	if n < 0 {
		return "", ErrNegativeNumber
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

	return strconv.Itoa(n), nil
}
