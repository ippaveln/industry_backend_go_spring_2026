package main

func reverseRunes(s string) string {
	sRunes := []rune(s)
	left := 0
	right := len(sRunes) - 1
	for left < right {
		sRunes[left], sRunes[right] = sRunes[right], sRunes[left]
		left++
		right--
	}
	return string(sRunes)
}
