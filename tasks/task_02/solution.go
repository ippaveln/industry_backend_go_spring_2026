package main

func reverseRunes(s string) string {
	runes := []rune(s)
	reversed := make([]rune, len(runes))
	for i := 0; i < len(runes); i++ {
		reversed[i] = runes[len(runes)-1-i]
	}
	return string(reversed)
}

