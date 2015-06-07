package parser

import (
	"container/list"
	"fmt"
	"github.com/kpmy/ypk/assert"
	"github.com/kpmy/ypk/halt"
	"io"
	"leaf/ebnf"
	"leaf/ebnf/tool"
	"leaf/meta/scanner"
	"os"
	"reflect"
)

const (
	enter  = -1
	must   = 1
	should = 0
)

type Parser interface {
	Compile(io.RuneReader)
}

type item struct {
	mod  int
	expr ebnf.Expression
}

type p0 struct {
	grammar ebnf.Grammar
	sc      scanner.Scanner
	sym     scanner.Symbol
	stack   *list.List
}

func (p *p0) push(i *item) {
	p.stack.PushFront(i)
}

func (p *p0) pop() {
	if p.stack.Len() > 1 {
		p.stack.Remove(p.stack.Front())
	}
}

func (p *p0) prev() (ret *item) {
	e := p.stack.Front()
	if e != nil {
		ee := e.Prev()
		if ee != nil {
			ret = ee.Value.(*item)
		}
	}
	return
}

func (p *p0) prePrev() (ret *item) {
	e := p.stack.Front()
	if e != nil {
		ee := e.Prev()
		if ee != nil {
			eee := ee.Prev()
			if eee != nil {
				ret = eee.Value.(*item)
			}
		}
	}
	return
}

func (p *p0) get() {
	p.sym = p.sc.Get()
}

func (p *p0) check(mode int, sym scanner.Symbol, msg string) (br bool) {
	if p.sym == sym {
		p.get()
	} else if mode == must {
		p.sc.Mark(msg)
		br = true
	}
	return
}

func (p *p0) do(mode int, _e ebnf.Expression) (br bool) {
	p.push(&item{mod: mode, expr: _e})
	switch e := _e.(type) {
	case ebnf.Sequence:
		fmt.Println("sequence")
		for _, i := range e {
			if p.do(must, i) {
				break
				br = true
			}
		}
	case *ebnf.Repetition:
		fmt.Println("repetition")
		p.do(should, e.Body)
	case ebnf.Alternative:
		fmt.Println("alternative")
		for _, i := range e {
			done := false
			for {
				ok := p.do(must, i)
				if !ok {
					done = true
					break
				}
			}
			br = !done
		}
	case *ebnf.Name:
		switch e.String {
		case "CRLF":
			p.get()
			br = p.check(mode, scanner.Newline, "new line?")
		case "LETTER":
			if p.sym == scanner.Char {
				ch := p.sc.Char()[0]
				br = (ch < 'A' || ch > 'Z' && ch < 'a' || ch > 'z')
				p.get()
			} else {
				br = true
			}
		default:
			prod := p.grammar[e.String]
			if prod != nil {
				br = p.do(mode, prod.Expr)

			} else {
				halt.As(101, mode, " @", e.String)
			}
		}
	case *ebnf.Token:
		fmt.Println("expected", e.String)
		buf := p.sc.Char()
		for buf != e.String {
			p.get()
			if p.sym == scanner.Char {
				buf += p.sc.Char()
			} else {
				break
			}
		}
		if buf != e.String {
			p.sc.Mark(e.String, " expected")
		} else {
			fmt.Println(e.String, " found")
		}
		p.get()
	default:
		halt.As(100, mode, reflect.TypeOf(e))
	}
	p.pop()
	return
}

func (p *p0) Compile(rd io.RuneReader) {
	p.sc.Init(rd)
	top := tool.Top(p.grammar)
	fmt.Println(top.Name.String)
	if p.do(enter, top.Expr) {
		p.sc.Mark("compilation failed")
	}
}

const l0 = "leaf0.ebnf"

func New() Parser {
	if f, err := os.Open(l0); err == nil {
		ret := &p0{}
		ret.grammar, err = ebnf.Parse(l0, f)
		assert.For(err == nil, 40)
		ret.sc = scanner.New()
		ret.stack = list.New()
		return ret
	} else {
		panic(err)
	}
}
