package parser

import (
	"fmt"
	"github.com/kpmy/ypk/assert"
	"leaf/scanner"
)

type Parser interface {
	Module() error
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
	if p.sym.Code != scanner.Null {
		fmt.Print("this ")
		fmt.Print("`" + p.sym.String() + "`")
	}
	p.sym = p.sc.Get()
	fmt.Print(" next ")
	fmt.Println("`" + p.sym.String() + "`")
	return p.sym
}

func (p *pr) init() {
	p.next()
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

func (p *pr) annotate() {
	assert.For(p.sym.Code == scanner.Lbrak, 20, "left brace expected here")
	p.next()
	key := true
	for {
		p.pass(scanner.Separator)
		if p.await(scanner.String) {
			if key {
				fmt.Println("KEY", p.sym.Str)
				p.next()
				key = false
				if p.await(scanner.Colon, scanner.Separator) {
					p.next()
				} else {
					p.sc.Mark("colon expected")
				}
			} else {
				fmt.Println("VALUE", p.sym.Str)
				p.next()
				key = true
				if p.await(scanner.Comma, scanner.Separator, scanner.Delimiter) {
					p.next()
				} else {
					if p.sym.Code == scanner.Rbrak {
						p.next()
						break
					} else {
						p.sc.Mark("comma or ] expected")
					}
				}

			}
		} else {
			p.pass(scanner.Separator)
			if p.sym.Code == scanner.Rbrak {
				p.next()
				break
			} else {
				p.sc.Mark("key : value or ] expected")
			}
		}
	}
}

func (p *pr) importDecl() {
	assert.For(p.sym.Code == scanner.Import, 20, "import section here")
	p.next()
	if p.sym.Code == scanner.Separator || p.sym.Code == scanner.Delimiter {
		for {
			if p.await(scanner.Ident, scanner.Separator, scanner.Delimiter) {
				fmt.Print("IMPORT ", p.sym.Str)
				p.next()
				if p.await(scanner.Becomes, scanner.Separator) {
					p.next()
					if p.await(scanner.Ident, scanner.Separator) {
						fmt.Println("FOR", p.sym.Str)
						p.next()
					} else {
						p.sc.Mark("module identifier expected")
					}
				}
				p.pass(scanner.Separator)
				if p.sym.Code == scanner.Lbrak {
					p.annotate()
				}
				if p.await(scanner.Comma, scanner.Separator) {
					p.next()
					fmt.Println()
				} else if p.sym.Code == scanner.Delimiter {
					break
				} else {
					p.sc.Mark("Comma or delimiter expected")
				}
			} else {
				p.sc.Mark("module identifier expected")
			}
		}
	} else {
		p.sc.Mark("separator expected")
	}
}

func (p *pr) expression() {
	p.sc.Mark("not implemented")
}

func (p *pr) constDecl() {
	assert.For(p.sym.Code == scanner.Const, 20, "const section here")
	p.next()
	if p.sym.Code == scanner.Separator || p.sym.Code == scanner.Delimiter {
		for {
			if p.await(scanner.Ident, scanner.Separator, scanner.Delimiter) {
				fmt.Print("CONST ", p.sym.Str)
				p.next()
				if p.sym.Code == scanner.Times {
					p.next()
					fmt.Println("exported")
				}
				if p.await(scanner.Delimiter, scanner.Separator) {
					// ATOM const
					fmt.Println("ATOM")
					p.next()
				} else if p.sym.Code == scanner.Equal {
					p.next()
					p.pass(scanner.Separator)
					p.expression()
				} else {
					p.sc.Mark("= expected")
				}
			} else {
				break
			}
		}
	} else {
		p.sc.Mark("separator expected")
	}
}

func (p *pr) Module() (err error) {
	fmt.Println("COMPILER")
	if p.await(scanner.Module, scanner.Separator, scanner.Delimiter) {
		p.next()
		if p.await(scanner.Ident, scanner.Separator) {
			fmt.Println("MODULE", p.sym.Str)
			p.next()
			p.pass(scanner.Separator)
			if p.sym.Code == scanner.Lbrak {
				p.annotate()
			}
			if p.await(scanner.Delimiter, scanner.Separator) {
				if p.await(scanner.Import, scanner.Delimiter, scanner.Separator) {
					p.importDecl()
				}
				for p.await(scanner.Const, scanner.Delimiter, scanner.Separator) {
					p.constDecl()
				}
			} else {
				p.sc.Mark("delimiter expected")
			}
		} else {
			p.sc.Mark("module name expected")
		}
	} else {
		p.sc.Mark("MODULE expected")
	}
	return
}

func ConnectTo(s scanner.Scanner) Parser {
	ret := &pr{sc: s}
	ret.init()
	return ret
}
