package scanner

import (
	"strconv"
	"unicode"
)

func Token2(r rune) string {
	return string([]rune{r})
}

func Token(r rune) string {
	if unicode.IsSpace(r) {
		return strconv.Itoa(int(r)) + "U"
	} else {
		return string([]rune{r})
	}
}
