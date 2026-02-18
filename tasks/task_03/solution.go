package main

import "fmt"

func fizzBuzz(n int) (string, error) {
	if n < 0 {
		return "", fmt.Errorf("Negative number!")
	}
	res := ""
	if n%3 == 0 {
		res += "Fizz"
	}
	if n%5 == 0 {
		res += "Buzz"
	}
	if res == "" {
		res = fmt.Sprintf("%d", n)
	}
	return res, nil
}
