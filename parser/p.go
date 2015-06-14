package parser

import (
	"fmt"
	"github.com/kpmy/ypk/assert"
	"leaf/ir"
	"leaf/scanner"
)

type Parser interface {
	Module() (*ir.Module, error)
}

type idgen struct {
	next int
}

func (i *idgen) nextID() (ret int) {
	ret = i.next
	i.next++
	return
}

type pr struct {
	sc   scanner.Scanner
	sym  scanner.Sym
	done bool
	t    target
	idgen
}

func (p *pr) next() scanner.Sym {
	p.done = true
	if p.sym.Code != scanner.Null {
		//		fmt.Print("this ")
		//		fmt.Print("`" + fmt.Sprint(p.sym) + "`")
	}
	p.sym = p.sc.Get()
	//	fmt.Print(" next ")
	fmt.Println("`" + fmt.Sprint(p.sym) + "`")
	return p.sym
}

func (p *pr) init() {
	p.next()
}

//expect is the most powerful step forward runner, breaks the compilation if unexpected sym found
func (p *pr) expect(sym scanner.Symbol, msg string, skip ...scanner.Symbol) {
	assert.For(p.done, 20)
	if !p.await(sym, skip...) {
		p.sc.Mark(msg)
	}
	p.done = false
}

//await runs for the sym through skip list, but may not find the sym
func (p *pr) await(sym scanner.Symbol, skip ...scanner.Symbol) bool {
	assert.For(p.done, 20)
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
	p.done = p.sym.Code != sym
	return p.sym.Code == sym
}

//pass runs through skip list
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

//run runs to the first sym through any other sym
func (p *pr) run(sym scanner.Symbol) {
	for p.next().Code != sym {
		if p.sc.Error() != nil {
			p.sc.Mark("not found")
			break
		}
	}
}

func (p *pr) ident() string {
	assert.For(p.sym.Code == scanner.Ident, 20, "identifier expected")
	//добавить валидацию идентификаторов
	return p.sym.Str
}

func (p *pr) is(sym scanner.Symbol) bool {
	return p.sym.Code == sym
}

func (p *pr) factor() {
	if p.is(scanner.Number) {
		p.next()
	} else {
		p.sc.Mark("not implemented")
	}
}

func (p *pr) term() {
	p.factor()
}

func (p *pr) simpleExpression() {
	p.term()
}

func (p *pr) expression() {
	p.simpleExpression()
}

func (p *pr) constDecl() {
	assert.For(p.sym.Code == scanner.Const, 20, "CONST block expected")
	p.next()
	for {
		if p.await(scanner.Ident, scanner.Delimiter, scanner.Separator) {
			p.next()
			if p.await(scanner.Equal, scanner.Separator) {
				p.next()
				p.pass(scanner.Separator)
				p.expression()
			} else if p.is(scanner.Delimiter) {
				p.next()
			}
		} else {
			break
		}

	}
}

func (p *pr) Module() (ret *ir.Module, err error) {
	p.expect(scanner.Module, "MODULE expected", scanner.Delimiter, scanner.Separator)
	p.next()
	p.expect(scanner.Ident, "module name expected", scanner.Separator)
	p.t.init(p.ident())
	p.next()
	p.pass(scanner.Separator, scanner.Delimiter)
	for p.await(scanner.Const, scanner.Delimiter, scanner.Separator) {
		p.constDecl()
	}
	p.expect(scanner.End, "no END", scanner.Delimiter, scanner.Separator)
	p.next()
	p.expect(scanner.Ident, "module name expected", scanner.Separator)
	if p.ident() != p.t.root.ModName {
		p.sc.Mark("module name does not match")
	}
	p.next()
	p.expect(scanner.Period, "end of module expected")
	ret = p.t.root
	return
}

func ConnectTo(s scanner.Scanner) Parser {
	assert.For(s != nil, 20)
	ret := &pr{sc: s}
	ret.init()
	return ret
}
