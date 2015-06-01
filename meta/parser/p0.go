package parser

import (
	"fmt"
	"github.com/kpmy/ypk/assert"
	"io"
	"leaf/ebnf"
	"leaf/ebnf/tool"
	"os"
)

type Parser interface {
	Compile(io.RuneReader)
}

type p0 struct {
	grammar ebnf.Grammar
}

func (p *p0) Compile(io.RuneReader) {}

const l0 = "leaf0.ebnf"

func New() Parser {
	if f, err := os.Open(l0); err == nil {
		ret := &p0{}
		ret.grammar, err = ebnf.Parse(l0, f)
		assert.For(err == nil, 40)
		g := tool.Transform(ret.grammar)
		fmt.Println(g)
		return ret
	} else {
		panic(err)
	}
}
