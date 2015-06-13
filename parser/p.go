package parser

import (
	"fmt"
	"github.com/kpmy/ypk/assert"
	"leaf/scanner"
)

type Parser interface {
	Module() error
}

type pr struct {
	sc  scanner.Scanner
	sym scanner.Sym
}

func (p *pr) next() scanner.Sym {
	if p.sym.Code != scanner.Null {
		//		fmt.Print("this ")
		//		fmt.Print("`" + fmt.Sprint(p.sym) + "`")
	}
	p.sym = p.sc.Get()
	//	fmt.Print(" next ")
	//	fmt.Println("`" + fmt.Sprint(p.sym) + "`")
	return p.sym
}

func (p *pr) init() {
	p.next()
}

func (p *pr) expect(sym scanner.Symbol, msg string, skip ...scanner.Symbol) {
	if !p.await(sym, skip...) {
		p.sc.Mark(msg)
	}
}

func (p *pr) await(sym scanner.Symbol, skip ...scanner.Symbol) bool {

	skipped := func() (ret bool) {
		for _, v := range skip {
			if v == p.sym.Code {
				ret = true
			}
		}
		return
	}

	for sym != p.sym.Code && skipped() {
		p.next()
	}

	return p.sym.Code == sym
}

func (p *pr) pass(skip ...scanner.Symbol) {
	skipped := func() (ret bool) {
		for _, v := range skip {
			if v == p.sym.Code {
				ret = true
			}
		}
		return
	}
	for skipped() {
		p.next()
	}
}

func (p *pr) Module() (err error) {
	fmt.Println("COMPILER")

	return
}

func ConnectTo(s scanner.Scanner) Parser {
	assert.For(s != nil, 20)
	ret := &pr{sc: s}
	ret.init()
	return ret
}
