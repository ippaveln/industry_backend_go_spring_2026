package main

import "strings"

func reverseRunes(s string) string {
	var builder strings.Builder
	runes := []rune(s)
	builder.Grow(len(runes))

	for i := len(runes) - 1; i >= 0; i-- {
		builder.WriteRune(runes[i])
	}

	return builder.String()
}
