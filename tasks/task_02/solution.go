package main

// Через свап местами
func reverseRunes(s string) string {
	rune_slice := []rune(s)
	for i := 0; i < len(rune_slice)/2; i++ {

		rune_slice[i], rune_slice[len(rune_slice)-1-i] = rune_slice[len(rune_slice)-1-i], rune_slice[i]
	}

	return string(rune_slice)
}

/*
Через создание слайса в который добавляем руны с конца из исходного слайса

func reverseRunes(s string) string{
	rune_slice := []rune(s)
	reversed := make([]rune,0,len(runes))
	for i := len(runes) - 1; i >= 0; i--{
		reversed = append(reversed, rune[i])
	}

	return string(reversed)
}


*/
