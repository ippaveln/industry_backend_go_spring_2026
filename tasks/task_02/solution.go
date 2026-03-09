package main

import (
	_ "fmt"
	"strings"
)

func reverseRunes(s string) string {
	if s == "" {
		return s
	}

	var runes = []rune(s)
	var sb strings.Builder

	for i := len(runes) - 1; i >= 0; i-- {
		sb.WriteString(string(runes[i]))
	}
	return sb.String()
}
