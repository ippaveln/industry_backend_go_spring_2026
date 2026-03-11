package main

import (
	"fmt"
	"strconv"
)

func fizzBuzz(n int) (string, error) {
	if n < 0 {
		return "", fmt.Errorf("This is a negative number!") 
	} else if n%3 == 0 && n%5 == 0 {
		return "FizzBuzz", nil
	} else if n%3 == 0 {
		return "Fizz", nil
	} else if n%5 == 0 {
		return "Buzz", nil
	}
	return strconv.Itoa(n), nil
}
