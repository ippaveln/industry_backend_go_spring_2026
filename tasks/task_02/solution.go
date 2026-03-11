package main

import "slices"

func reverseRunes(s string) string {
	if s == "" {
		return s
	}

	runes := []rune(s)
	slices.Reverse(runes)
	
	return string(runes)
}
