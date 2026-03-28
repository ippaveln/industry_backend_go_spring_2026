package main

func reverseRunes(s string) string {
	runes := []rune(s)
	n := len(runes)
	for i := 0; i < n/2; i++ {
		runes[i], runes[n-1-i] = runes[n-1-i], runes[i]
	}
	return string(runes)
}

//// С использованием strings.Builder и defer
//func reverseRunes(s string) string {
//	var sb strings.Builder
//	writeToBuilder(s, &sb)
//	return sb.String()
//}
//
//func writeToBuilder(s string, sb *strings.Builder) {
//	var cnt int
//	for _, r := range s {
//		cnt++
//		defer sb.WriteRune(r)
//	}
//	sb.Grow(cnt)
//}
