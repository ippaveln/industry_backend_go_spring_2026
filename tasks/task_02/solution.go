package main

type stringBuilder string

func (str *stringBuilder) addString(char rune) {
	*str = *str + stringBuilder(string(char))
}

func (str *stringBuilder) reverse() {
	for _, char := range *str {
		defer str.addString(char)
	}
	*str = ""
}

func reverseRunes(s string) string {
	str := stringBuilder(s)
	str.reverse()
	return string(str)
}
