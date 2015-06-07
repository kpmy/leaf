package scanner

import (
	"errors"
	"fmt"
	"github.com/kpmy/ypk/assert"
	"io"
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
	wrong   Symbol = 0
	Null    Symbol = -1
	Newline        = 101
	Char           = 20
)

func (s Symbol) String() (ret string) {
	switch s {
	case wrong:
		ret = "wrong sym"
	case Null:
		ret = "null"
	case Newline:
		ret = "newline"
	case Char:
		ret = "char"
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
	Mark(...interface{})
	Error() error
	Eot() bool
	Char() string
}

type sc struct {
	rd      io.RuneReader
	ch      rune
	eof     int
	pos     int64
	id      string
	val     *big.Int
	err     error
	errPos  int64
	lastSym Symbol
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
	for !s.eot() && s.ch <= ' ' && !(s.ch == '\n' || s.ch == '\r') {
		s.read()
	}
}

func (s *sc) Mark(msg ...interface{}) {
	log.Println("pos", s.pos, "::", fmt.Sprint(msg...))
	s.errPos = s.pos
	s.err = errors.New(fmt.Sprint(msg...))
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
func (s *sc) Char() string  { return string([]rune{s.ch}) }

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
		switch {
		case s.ch == '\r' || s.ch == '\n':
			s.read()
			if s.ch == '\n' {
				s.read()
				sym = Newline
			} else {
				sym = Newline
			}
		case s.ch == '(':
			s.read()
			if s.ch == '*' {
				sym = Null
				s.comment()
			} else {
				sym = Char
				s.ch = '('
			}
		default:
			s.char = s.ch
			s.read()
			sym = Char
		}
		if sym != Null || s.err != nil || s.eot() {
			break
		}
	}
	s.lastSym = sym
	fmt.Println("sym ::", sym)
	return
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
