package main

func reverseRunes(s string) string {
	runeStr := []rune(s)
	for l, r := 0, len(runeStr) - 1; l < r; {
		runeStr[l], runeStr[r] = runeStr[r], runeStr[l]
		l++
		r--
	}
	return string(runeStr)
}
