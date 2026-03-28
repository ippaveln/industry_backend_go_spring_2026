package main

import (
	"errors"
	"strconv"
	"strings"
)

const (
	numFizz = 3
	numBuzz = 5

	strFizz = "Fizz"
	strBuzz = "Buzz"
)

var (
	errNegativeNum = errors.New("passed number must not be negative")
)

func fizzBuzz(n int) (s string, err error) {
	if n < 0 {
		return "", errNegativeNum
	}

	var sb strings.Builder
	sb.Grow(len(strFizz) + len(strBuzz))
	if n%numFizz == 0 {
		sb.WriteString(strFizz)
	}
	if n%numBuzz == 0 {
		sb.WriteString(strBuzz)
	}

	if s = sb.String(); s == "" {
		return strconv.Itoa(n), nil
	}
	return s, nil
}
