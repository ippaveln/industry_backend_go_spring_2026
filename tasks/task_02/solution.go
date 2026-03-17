package main

func reverseRunes(s string) string {
	var result string

	runes := []rune(s)
	for i := len(runes) - 1; i >= 0; i-- {
		result += string(runes[i])
	}

	return result
}
