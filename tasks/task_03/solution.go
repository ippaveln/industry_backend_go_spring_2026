package main

import (
	"errors"
	"strconv"
)

func fizzBuzz(n int) (string, error) {
	if n < 0 {
		return "", errors.New("input must be a positive integer")
	}
	if n%3 == 0 && n%5 == 0 {
		return "FizzBuzz", nil
	} else if n%5 == 0 {
		return "Buzz", nil
	} else if n%3 == 0 {
		return "Fizz", nil
	}

	return strconv.Itoa(n), nil
}
