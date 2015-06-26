package lss

import (
	"strconv"
	"unicode"
)

var keyTab map[string]Symbol

func keyByTab(s Symbol) (ret string) {
	for k, v := range keyTab {
		if v == s {
			ret = k
		}
	}
	return
}

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

func init() {
	keyTab = map[string]Symbol{"MODULE": Module,
		"DEFINITION": Definition,
		"END":        End,
		"DO":         Do,
		"WHILE":      While,
		"ELSIF":      Elsif,
		"IMPORT":     Import,
		"CONST":      Const,
		"OF":         Of,
		"PRE":        Pre,
		"POST":       Post,
		"PROCEDURE":  Proc,
		"VAR":        Var,
		"BEGIN":      Begin,
		"CLOSE":      Close,
		"IF":         If,
		"THEN":       Then,
		"REPEAT":     Repeat,
		"UNTIL":      Until,
		"ELSE":       Else,
		"TRUE":       True,
		"FALSE":      False,
		"NIL":        Nil,
		"INF":        Inf,
		"CHOOSE":     Choose,
		"OR":         Opt,
		"INFIX":      Infix}
}
