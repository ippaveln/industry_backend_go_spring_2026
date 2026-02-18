package main

func reverseRunes(s string) string {
	runes := []rune(s)
	var res []rune
	for i := len(runes) - 1; i >= 0; i-- {
		res = append(res, runes[i])
	}
	return string(res)
}
