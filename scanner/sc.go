package scanner

import (
	"fmt"
	"github.com/kpmy/ypk/assert"
	"github.com/kpmy/ypk/halt"
	"io"
	"log"
	"strconv"
	"strings"
	"unicode"
)

type Symbol int

const (
	Null = iota
	Period
	Delimiter
	Separator
	Ident
	String
	Becomes
	Number
	UpTo

	Lbrak
	Rbrak
	Colon
	Comma
	Times
	Equal
	Nequal
	Minus
	Plus
	Lparen
	Rparen
	And
	Or
	Not
	Lbrace
	Rbrace
	Leq
	Lss
	Geq
	Gtr
	Arrow
	Div
	Divide
	Mod

	Module
	End
	Do
	While
	Elsif
	Import
	Const
	Type
	Of
	To
	This
	In
	Out
	Io
	Pre
	Post
	Proc
	Var
	Begin
	Close
	Match
	Case
	If
	Then
	Repeat
	Until
	Else
	True
	False
	Nil
	With
)

var keyTab map[string]Symbol

func init() {
	keyTab = map[string]Symbol{"MODULE": Module,
		"END":       End,
		"DO":        Do,
		"WHILE":     While,
		"ELSIF":     Elsif,
		"IMPORT":    Import,
		"CONST":     Const,
		"TYPE":      Type,
		"OF":        Of,
		"TO":        To,
		"THIS":      This,
		"IN":        In,
		"OUT":       Out,
		"IO":        Io,
		"PRE":       Pre,
		"POST":      Post,
		"PROCEDURE": Proc,
		"VAR":       Var,
		"BEGIN":     Begin,
		"CLOSE":     Close,
		"MATCH":     Match,
		"CASE":      Case,
		"IF":        If,
		"THEN":      Then,
		"REPEAT":    Repeat,
		"UNTIL":     Until,
		"ELSE":      Else,
		"TRUE":      True,
		"FALSE":     False,
		"NIL":       Nil,
		"WITH":      With}
}

func keyByTab(s Symbol) (ret string) {
	for k, v := range keyTab {
		if v == s {
			ret = k
		}
	}
	return
}

func (s Symbol) String() (ret string) {
	switch s {
	case Module, End, Do, While, Elsif, Import, Const, Type, Of, To, This, In, Out, Io, Pre, Post, Proc, Var, Begin, Close, Match, If, Case, Then, Repeat, Until, Else, True, False, Nil, With:
		ret = keyByTab(s)
	case Null:
		ret = "null"
	case Period:
		ret = "."
	case Delimiter:
		ret = "delimiter"
	case Ident:
		ret = "identifier"
	case Separator:
		ret = "separator"
	case String:
		ret = "string"
	case Becomes:
		ret = ":="
	case Lbrak:
		ret = "["
	case Rbrak:
		ret = "]"
	case Lbrace:
		ret = "{"
	case Rbrace:
		ret = "}"
	case Colon:
		ret = ":"
	case Comma:
		ret = ","
	case Times:
		ret = "*"
	case Equal:
		ret = "="
	case Minus:
		ret = "-"
	case Nequal:
		ret = "#"
	case Geq:
		ret = ">="
	case Gtr:
		ret = ">"
	case Leq:
		ret = "<="
	case Lss:
		ret = "<"
	case Lparen:
		ret = "("
	case Rparen:
		ret = ")"
	case And:
		ret = "&"
	case Or:
		ret = "|"
	case Not:
		ret = "~"
	case Plus:
		ret = "+"
	case Arrow:
		ret = "^"
	case Div:
		ret = "//"
	case Divide:
		ret = "/"
	case Mod:
		ret = "%"
	case Number:
		ret = "num"
	case UpTo:
		ret = ".."
	default:
		ret = fmt.Sprint("sym [", strconv.Itoa(int(s)), "]")
	}
	return
}

type Sym struct {
	Code Symbol
	Str  string
}

func (v Sym) String() (ret string) {
	switch v.Code {
	case Ident:
		ret = fmt.Sprint(`@` + v.Str)
	case Delimiter:
		ret = fmt.Sprint(";")
	case Separator:
		ret = fmt.Sprint(" ")
	case String:
		ret = fmt.Sprint(`"` + v.Str + `"`)
	case Number:
		ret = fmt.Sprint(v.Str)
	default:
		ret = fmt.Sprint(v.Code)
	}
	return
}

type Scanner interface {
	Get() Sym
	Error() error
	Mark(...interface{})
}

func ConnectTo(r io.RuneReader) Scanner {
	ret := &sc{rd: r}
	ret.init()
	return ret
}

type sc struct {
	rd  io.RuneReader
	err error
	pos int

	ch rune
}

func (s *sc) Error() error { return s.err }

func (s *sc) Mark(msg ...interface{}) {
	log.Println("at pos ", s.pos, " ", fmt.Sprintln(msg...))
	halt.As(100, "at pos ", s.pos, " ", fmt.Sprintln(msg...))
}

func (s *sc) next() rune {
	//	fmt.Print(Token2(s.ch))
	read := 0
	s.ch, read, s.err = s.rd.ReadRune()
	if s.err == nil {
		s.pos += read
	}
	return s.ch
}

func (s *sc) ident() (sym Sym) {
	assert.For(unicode.IsLetter(s.ch), 20, "character expected")
	var buf []rune
	for {
		buf = append(buf, s.ch)
		s.next()
		if s.err != nil || !(unicode.IsLetter(s.ch) || unicode.IsDigit(s.ch)) {
			break
		}
	}
	if s.err == nil {
		sym.Str = string(buf)
		if sym.Code = keyTab[sym.Str]; sym.Code == Null {
			sym.Code = Ident
		}
	} else {
		halt.As(100, "error while ident read")
	}
	return
}

func (s *sc) comment() {
	assert.For(s.ch == '*', 20, "expected * ", "got ", Token(s.ch))
	for {
		for s.err == nil && s.ch != '*' {
			if s.ch == '(' {
				if s.next() == '*' {
					s.comment()
				}
			} else {
				s.next()
			}
		}
		for s.err == nil && s.ch == '*' {
			s.next()
		}
		if s.err != nil || s.ch == ')' {
			break
		}
	}
	if s.err == nil {
		s.next()
	} else {
		s.Mark("unclosed comment")
	}
}

func (s *sc) str() string {
	assert.For(s.ch == '"' || s.ch == '\'', 20, "quote expected")
	var buf []rune
	ending := s.ch
	s.next()
	for ; s.err == nil && s.ch != ending; s.next() {
		buf = append(buf, s.ch)
	}
	if s.err == nil {
		s.next()
	} else {
		halt.As(100, "string expected")
	}
	return string(buf)
}

const dec = "0123456789"
const hex = dec + "ABCDEF"
const non = "01234WXYZ"
const tri = "-0+"
const modifier = "BHNTU"

func Is(pattern string, x rune) bool {
	return strings.ContainsRune(pattern, x)
}

//first char always 0..9
func (s *sc) num() (sym Sym) {
	assert.For(unicode.IsDigit(s.ch), 20, "digit expected")
	var buf []rune
	for {
		buf = append(buf, s.ch)
		s.next()
		if s.err != nil || !(s.ch == '.' || Is(hex, s.ch) || Is(non, s.ch) || Is(tri, s.ch)) {
			break
		}
	}
	if Is(modifier, s.ch) {
		buf = append(buf, s.ch)
		s.next()
	}
	if s.err == nil {
		sym.Code = Number
		sym.Str = string(buf)
	} else {
		halt.As(100, "error reading number")
	}
	return
}

func (s *sc) Get() (sym Sym) {
	for {
		switch s.ch {
		case '.':
			if s.next() == '.' {
				s.next()
				sym.Code = UpTo
			} else {
				sym.Code = Period
			}
		case '(':
			if s.next() == '*' {
				s.comment()
			} else {
				sym.Code = Lparen
			}
		case ')':
			s.next()
			sym.Code = Rparen
		case '\r', '\n', ';':
			for ; s.ch == '\n' || s.ch == '\r' || s.ch == ';'; s.next() {
			}
			sym.Code = Delimiter
		case ' ', '\t':
			for ; s.ch == ' ' || s.ch == '\t'; s.next() {
			}
			sym.Code = Separator
		case '[':
			sym.Code = Lbrak
			s.next()
		case ']':
			sym.Code = Rbrak
			s.next()
		case '"', '\'':
			sym.Str = s.str()
			sym.Code = String
		case ':':
			if s.next() == '=' {
				s.next()
				sym.Code = Becomes
			} else {
				sym.Code = Colon
			}
		case ',':
			sym.Code = Comma
			s.next()
		case '*':
			sym.Code = Times
			s.next()
		case '=':
			sym.Code = Equal
			s.next()
		case '-':
			sym.Code = Minus
			s.next()
		case '#':
			sym.Code = Nequal
			s.next()
		case '<':
			if s.next() == '=' {
				s.next()
				sym.Code = Leq
			} else {
				sym.Code = Lss
			}
		case '>':
			if s.next() == '=' {
				s.next()
				sym.Code = Geq
			} else {
				sym.Code = Gtr
			}
		case '&':
			sym.Code = And
			s.next()
		case '|':
			sym.Code = Or
			s.next()
		case '~':
			sym.Code = Not
			s.next()
		case '+':
			sym.Code = Plus
			s.next()
		case '{':
			sym.Code = Lbrace
			s.next()
		case '}':
			sym.Code = Rbrace
			s.next()
		case '^':
			sym.Code = Arrow
			s.next()
		case '/':
			if s.next() == '/' {
				s.next()
				sym.Code = Div
			} else {
				sym.Code = Divide
			}
		case '%':
			sym.Code = Mod
			s.next()
		default:
			switch {
			case unicode.IsLetter(s.ch):
				sym = s.ident()
			case unicode.IsDigit(s.ch):
				sym = s.num()
			default:
				halt.As(100, "unhandled ", Token(s.ch))
				s.next()
			}
		}
		if s.err != nil || sym.Code != Null {
			break
		}
	}
	return
}

func (s *sc) init() {
	s.pos = 0
	s.next()
}
