package main

import (
	"fmt"
)

func fizzBuzz(n int) (string, error) {
	resultString := ""
	if n < 0 {
		return resultString, fmt.Errorf("negative number")
	}
	if n%3 == 0 {
		resultString += "Fizz"
	}
	if n%5 == 0 {
		resultString += "Buzz"
	}
	if resultString == "" {
		return fmt.Sprintf("%d", n), nil
	}
	return resultString, nil
}
