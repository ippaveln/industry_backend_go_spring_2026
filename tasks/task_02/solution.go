package main

func reverseRunes(s string) string {
	runes := []rune(s)

	for i := 0; i < len(runes)/2; i++ {
		var t rune
		t = runes[i]
		runes[i] = runes[len(runes)-1-i]
		runes[len(runes)-1-i] = t
	}

	return string(runes)
}
