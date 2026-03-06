package main

import (
	"errors"
	"strconv"
)

func fizzBuzz(n int) (string, error) {
	switch {
	case n < 0:
		return "", errors.New("n < 0")
	case n%15 == 0:
		return "FizzBuzz", nil
	case n%3 == 0:
		return "Fizz", nil
	case n%5 == 0:
		return "Buzz", nil
	}
	return strconv.Itoa(n), nil
}
