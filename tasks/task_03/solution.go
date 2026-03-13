package main

import (
	"errors"
	"strconv"
)

// - если `n` делится на 3 без остатка — вернуть `"Fizz"`
// - если `n` делится на 5 без остатка — вернуть `"Buzz"`
// - если `n` делится и на 3, и на 5 — вернуть `"FizzBuzz"`
// - если `n` не делится ни на 3, ни на 5 — вернуть строковое представление числа (например, для `7` вернуть `"7"`)
// - если `n` отрицательное — вернуть пустую строку `""` и ошибку (`error != nil`)

func fizzBuzz(n int) (string, error) {
	if n < 0 {
		return "", errors.ErrUnsupported
	}

	if n%15 == 0 {
		return "FizzBuzz", nil
	} else if n%3 == 0 {
		return "Fizz", nil
	} else if n%5 == 0 {
		return "Buzz", nil
	} else {
		return strconv.Itoa(n), nil
	}
}
