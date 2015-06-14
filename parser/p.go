package parser

import (
	"github.com/kpmy/ypk/assert"
	"leaf/ir"
	"leaf/scanner"
)

type Parser interface {
	Module() (*ir.Module, error)
}

type pr struct {
	sc   scanner.Scanner
	sym  scanner.Sym
	done bool
	t    target
}

func (p *pr) next() scanner.Sym {
	p.done = true
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
	p.done = false
	return p.sym.Code == sym
}

//pass runs through skip list
func (p *pr) pass(skip ...scanner.Symbol) {
	assert.For(p.done, 20)
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
	p.done = false
}

//run runs to the first sym through any other sym
func (p *pr) run(sym scanner.Symbol) {
	assert.For(p.done, 20)
	for p.next().Code != sym {
		if p.sc.Error() != nil {
			p.sc.Mark("not found")
			break
		}
	}
	p.done = false
}

func (p *pr) ident() string {
	assert.For(p.sym.Code == scanner.Ident, 20, "identifier expected")
	//добавить валидацию идентификаторов
	return p.sym.Str
}

func (p *pr) Module() (ret *ir.Module, err error) {
	p.expect(scanner.Module, "MODULE expected", scanner.Delimiter, scanner.Separator)
	p.next()
	p.expect(scanner.Ident, "module name expected", scanner.Separator)
	p.t.init(p.ident())
	p.next()
	p.run(scanner.End)
	p.next()
	p.expect(scanner.Ident, "module name expected", scanner.Separator)
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
