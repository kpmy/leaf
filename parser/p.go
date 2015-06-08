package parser

import (
	"leaf/scanner"
)

type Parser interface {
	Module()
}

type Type struct{}
type Proc struct{}
type Value struct{}

var entries map[string]interface{}

func init() {
	entries = map[string]interface{}{"SET": Type{},
		"MAP":     Type{},
		"LIST":    Type{},
		"POINTER": Type{},
		"STRING":  Type{},
		"ATOM":    Type{},
		"BOOLEAN": Type{},
		"TRILEAN": Type{},
		"INTEGER": Type{},
		"REAL":    Type{},
		"CHAR":    Type{},

		"NIL":   Value{},
		"TRUE":  Value{},
		"FALSE": Value{},

		"LEN": Proc{},
		"NEW": Proc{}}
}

type pr struct {
	sc  scanner.Scanner
	sym scanner.Sym
}

func (p *pr) next() scanner.Sym {
	p.sym = p.sc.Get()
	return p.sym
}

func (p *pr) init() {

}

func (p *pr) Module() error {

}

func ConnectTo(s scanner.Scanner) Parser {
	ret := &pr{sc: s}
	ret.init()
	return ret
}
