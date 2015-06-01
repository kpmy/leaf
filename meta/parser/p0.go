package parser

import (
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

type Parser interface {
	Compile(io.RuneReader)
}

type p0 struct {
	grammar ebnf.Grammar
	sc      scanner.Scanner
	sym     scanner.Symbol
}

func (p *p0) get() {
	p.sym = p.sc.Get()
}

func (p *p0) top(e ebnf.Expression) {
	var expect func(ebnf.Expression) bool
	expect = func(_e ebnf.Expression) bool {
		switch e := _e.(type) {
		case *ebnf.Token:
			p.get()
			fmt.Println(p.sym, p.sc.Id(), e.String)
			switch {
			case p.sym == scanner.Ident:
				return p.sc.Id() == e.String
			default:
				halt.As(101, p.sym)
			}
		default:
			halt.As(100, reflect.TypeOf(e))
		}
		panic(0)
	}
	fn := tool.Filter(p.grammar, p.sc.(tool.Tokenizer))
	if err := fn(e, expect); err != nil {
		p.sc.Mark(fmt.Sprint("expectation failed: ", err))
	}
}

func (p *p0) Compile(rd io.RuneReader) {
	p.sc.Init(rd)
	p.top(tool.Top(p.grammar).Expr)
}

const l0 = "leaf0.ebnf"

func New() Parser {
	if f, err := os.Open(l0); err == nil {
		ret := &p0{}
		ret.grammar, err = ebnf.Parse(l0, f)
		assert.For(err == nil, 40)
		ret.sc = scanner.New()
		return ret
	} else {
		panic(err)
	}
}
