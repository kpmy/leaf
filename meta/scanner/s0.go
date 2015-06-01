package scanner

import (
	"errors"
	"fmt"
	"github.com/kpmy/ypk/assert"
	"io"
	"leaf/ebnf"
	"leaf/ebnf/tool"
	"log"
	"math/big"
	"strconv"
)

type Symbol int

var keyTab map[string]Symbol

const (
	NewLine = "\r\n"
)

const (
	wrong     Symbol = 0
	Null      Symbol = -1
	Ident            = 31
	Semicolon        = 52
	CRLF             = 101
	Char             = 20
	Times            = 1
	Neq              = 10
	Lparen           = 28
	Rparen           = 44
	Plus             = 6
	Minus            = 7
	Comma            = 40
	Period           = 18
	Lbrak            = 29
	And              = 5
	Not              = 27
	Rbrak            = 45
	Int              = 21
	Becomes          = 42
	Colon            = 41
	Leq              = 12
	Lss              = 11
	Gtr              = 13
	Geq              = 14
	Eql              = 9
	Begin            = 67
	End              = 53
)

func (s Symbol) String() (ret string) {
	switch s {
	case wrong:
		ret = "wrong sym"
	case Null:
		ret = "null"
	case Ident:
		ret = "ident"
	case Period:
		ret = "period"
	default:
		for k, v := range keyTab {
			if v == s {
				ret = k
			}
		}
		if ret == "" {
			ret = fmt.Sprint(strconv.Itoa(int(s)), " fixme")
		}
	}
	return
}

type Scanner interface {
	Init(io.RuneReader)
	Get() Symbol
	Mark(string)
	Id() string
	Val() *big.Int
	Error() error
	Eot() bool
}

type sc struct {
	rd     io.RuneReader
	ch     rune
	eof    int
	pos    int64
	id     string
	val    *big.Int
	err    error
	errPos int64
}

//TODO non-unicode, fix it
func (s *sc) read() {
	//buf := make([]byte, 1)
	ch, read, err := s.rd.ReadRune()
	if err != nil {
		s.eof++
		s.ch = 0
	} else {
		s.eof = 0
		s.ch = ch
		s.pos += int64(read)
		fmt.Println("read ", s.pos, ":", s.ch)
	}
}

func (s *sc) Eot() bool {
	return s.eof > 1
}

func (s *sc) eot() bool {
	return s.Eot()
}
func (s *sc) skipWhite() {
	fmt.Println("skip white")
	for !s.eot() && s.ch <= ' ' && !(s.ch == '\n' || s.ch == '\r') {
		s.read()
	}
	fmt.Println("white skipped")
}

func (s *sc) Mark(msg string) {
	log.Println("pos", s.pos, "::", msg)
	s.errPos = s.pos
	s.err = errors.New(msg)
}

func (s *sc) comment() {
	for {
		for {
			s.read()
			for s.ch == '(' {
				s.read()
				if s.ch == '*' {
					s.comment()
				}
			}
			if s.ch == '*' || s.eot() {
				break
			}
		}
		for {
			s.read()
			if s.ch != '*' || s.eot() {
				break
			}
		}
		if s.ch == ')' || s.eot() {
			break
		}
	}
	if !s.eot() {
		s.read()
	} else {
		s.Mark("comment not terminated")
	}
}

func (s *sc) Id() string    { return s.id }
func (s *sc) Val() *big.Int { return s.val }
func (s *sc) Error() error  { return s.err }

func (s *sc) ident() (sym Symbol) {
	fmt.Println("get ident")
	buf := make([]rune, 0)
	for {
		buf = append(buf, s.ch)
		s.read()
		if (s.ch < '0' || s.ch > '9') && (s.ch < 'A' || s.ch > 'Z') && (s.ch < 'a' || s.ch > 'z') {
			break
		}
	}
	s.id = string(buf)
	sym = keyTab[s.id]
	if sym == wrong {
		sym = Ident
	}
	fmt.Println("ident", sym, s.id)
	return
}

func (s *sc) number() (sym Symbol) {
	sym = Int
	var buf []byte
	for {
		buf = append(buf, byte(s.ch))
		s.read()
		if s.ch < '0' || s.ch > '9' {
			break
		}
	}
	assert.For(s.val.UnmarshalText(buf) == nil, 60, buf)
	return
}

//  Get from areas
// \r
// \n
// !"#$%&'()*+,-./0123456789:;<=>?@
// ABCDEFGHIJKLMNOPQRSTUVWXYZ
// [\]^_`
// abcdefghijklmnopqrstuvwxyz
// {|}~
func (s *sc) Get() (sym Symbol) {
	assert.For(!s.eot(), 20)
	for {
		s.skipWhite()
		if s.ch < ' ' {
			fmt.Println("raw", s.pos, s.ch)
		} else {
			fmt.Println("scan", s.pos, string([]rune{s.ch}))
		}
		switch {
		case s.ch == ';':
			s.read()
			sym = Semicolon
		case s.ch == '\r' || s.ch == '\n':
			s.read()
			if s.ch == '\n' {
				s.read()
				sym = CRLF
			} else {
				sym = CRLF
			}
		case s.ch < 'A':
			switch {
			case s.ch < '0':
				switch {
				case s.ch == '"':
					s.read()
					for {
						s.read()
						if s.ch == '"' || s.eot() {
							break
						}
						s.read()
						sym = Char
					}
				case s.ch == '#':
					s.read()
					sym = Neq
				case s.ch == '&':
					s.read()
					sym = And
				case s.ch == '(':
					s.read()
					if s.ch == '*' {
						sym = Null
						s.comment()
					} else {
						sym = Lparen
					}
				case s.ch == ')':
					s.read()
					sym = Rparen
				case s.ch == '*':
					s.read()
					sym = Times
				case s.ch == '+':
					s.read()
					sym = Plus
				case s.ch == '-':
					s.read()
					sym = Minus
				case s.ch == ',':
					s.read()
					sym = Comma
				case s.ch == '.':
					s.read()
					sym = Period
				default: // ! $ %
					s.read()
					sym = Null
				}
			case s.ch < ':':
				sym = s.number()
			case s.ch == ':':
				s.read()
				if s.ch == '=' {
					s.read()
					sym = Becomes
				} else {
					sym = Colon
				}
			case s.ch == '<':
				s.read()
				if s.ch == '=' {
					s.read()
					sym = Leq
				} else {
					sym = Lss
				}
			case s.ch == '=':
				s.read()
				sym = Eql
			case s.ch == '>':
				s.read()
				if s.ch == '=' {
					s.read()
					sym = Geq
				} else {
					sym = Gtr
				}
			default: // ? @
				s.read()
				sym = Null
			}
		case s.ch < '[':
			sym = s.ident()
		case s.ch < 'a':
			if s.ch == '[' {
				sym = Lbrak
			} else if s.ch == ']' {
				sym = Rbrak
			} else { // _ ` ^
				sym = Null
			}
			s.read()
		case s.ch < '{':
			sym = s.ident()
		default:
			if s.ch == '~' {
				sym = Not
			} else { // { } |
				sym = Null
			}
			s.read()
		}
		if sym != Null || s.err != nil || s.eot() {
			break
		}
	}
	fmt.Println("sym ::", sym)
	return
}

func (s *sc) Apply(ebnf.Expression, tool.Applicator) error {
	return nil
}

func (s *sc) Init(rd io.RuneReader) {
	s.rd = rd
	s.read()
}

func New() Scanner {
	return &sc{val: big.NewInt(0)}
}

func init() {
	keyTab = make(map[string]Symbol)
}
