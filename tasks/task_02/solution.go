package main

func reverseRunes(s string) string {
	runes := []rune(s)
	reversed := make([]rune, len(runes))

	for i := len(runes) - 1; i >= 0; i-- {
		reversed[len(runes)-i-1] = runes[i]
	}

	return string(reversed)
}
