package parser

import (
	"fmt"
	"github.com/kpmy/ypk/assert"
	"leaf/ir"
	"leaf/scanner"
	"leaf/target"
)

type Parser interface {
	Module() error
}

type Type struct {
	typ ir.Type
}

type Proc struct{}
type Value struct{}

var entries map[string]interface{}

func init() {
	entries = map[string]interface{}{"SET": Type{typ: ir.Set},
		"MAP":     Type{typ: ir.Map},
		"LIST":    Type{typ: ir.List},
		"POINTER": Type{typ: ir.Pointer},
		"STRING":  Type{typ: ir.String},
		"ATOM":    Type{typ: ir.Atom},
		"BOOLEAN": Type{typ: ir.Boolean},
		"TRILEAN": Type{typ: ir.Trilean},
		"INTEGER": Type{typ: ir.Integer},
		"REAL":    Type{typ: ir.Real},
		"CHAR":    Type{typ: ir.Char},

		"LEN": Proc{},
		"NEW": Proc{}}
}

type pr struct {
	sc  scanner.Scanner
	sym scanner.Sym
	tg  target.Target
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

func (p *pr) factor() {
	switch p.sym.Code {
	case scanner.Number:
		fmt.Println(p.sym.Str)
		p.next()
	case scanner.String:
		fmt.Println(p.sym.Str)
		p.next()
	case scanner.True:
		fmt.Println("TRUE")
		p.next()
	case scanner.False:
		fmt.Println("FALSE")
		p.next()
	case scanner.Nil:
		fmt.Println("NIL")
		p.next()
	default:
		p.sc.Mark("not a factor")
	}
}

func (p *pr) term() {
	p.factor()
}

func (p *pr) simpleExpr() {
	switch p.sym.Code {
	case scanner.Number:
		p.term()
	case scanner.Minus:
		fmt.Print("-")
		p.next()
		p.term()
	case scanner.Plus:
		fmt.Print("+")
		p.next()
		p.term()
	default:
		p.term()
	}
}

func (p *pr) expression() {
	p.simpleExpr()
	switch p.sym.Code {
	case scanner.Delimiter: //do nothing
	default:
		p.sc.Mark("not implemented")
	}
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

func (p *pr) listTyp() {
	assert.For(p.sym.Code == scanner.Ident, 20, "identifier expected")
	t, ok := entries[p.sym.Str].(Type)
	assert.For(ok && t.typ == ir.List, 21, "list expected here")
	p.next()
	if p.await(scanner.Number, scanner.Separator) {
		fmt.Println("size ", p.sym.Str)
		p.next()
	}
	if p.await(scanner.Of, scanner.Separator) {
		p.next()
		fmt.Println("OF")
		if p.await(scanner.Ident, scanner.Separator) {
			p.typ()
		}
		if p.await(scanner.With, scanner.Separator) {
			//list with predefined fields
			p.next()
			if p.await(scanner.Ident, scanner.Separator, scanner.Delimiter) {
				for {
					p.field()
					if p.await(scanner.Delimiter, scanner.Separator) {
						if p.await(scanner.End, scanner.Separator, scanner.Delimiter) {
							break
						} else if p.sym.Code == scanner.Ident {
							//очередное поле
						} else {
							p.sc.Mark("identifier expected")
						}
					} else {
						p.sc.Mark("map delimiter expected")
					}
				}
				if p.await(scanner.End, scanner.Separator, scanner.Delimiter) {
					p.next()
				} else {
					p.sc.Mark("END of list expected")
				}
			} else {
				p.sc.Mark("incorrect list definition")
			}
		}
	} else {
		p.sc.Mark("incorrect list definition")
	}
}

func (p *pr) procTyp() {
	assert.For(p.sym.Code == scanner.Proc, 20, "procedure expected here")
	p.next()
	p.pass(scanner.Delimiter, scanner.Separator)
	method := false
	stop := false
	for !stop {
		switch p.sym.Code {
		case scanner.This:
			p.next()
			if !method {
				method = true
				if p.await(scanner.Ident, scanner.Separator) {
					fmt.Println("THIS " + p.sym.Str)
					p.next()
					if p.await(scanner.Ident, scanner.Separator) {
						fmt.Println(p.sym.Str)
						p.typ()
					} else {
						p.sc.Mark("type name expected")
					}
				} else {
					p.sc.Mark("identifier expected")
				}
			} else {
				p.sc.Mark("THIS already exists")
			}
		case scanner.In:
			p.next()
			if p.await(scanner.Ident, scanner.Separator) {
				fmt.Println("IN " + p.sym.Str)
				p.next()
				if p.await(scanner.Ident, scanner.Separator) {
					fmt.Println(p.sym.Str)
					p.typ()
				} else {
					p.sc.Mark("type name expected")
				}
			} else {
				p.sc.Mark("identifier expected")
			}
		case scanner.Out:
			p.next()
			if p.await(scanner.Ident, scanner.Separator) {
				fmt.Println("OUT " + p.sym.Str)
				p.next()
				if p.await(scanner.Ident, scanner.Separator) {
					fmt.Println(p.sym.Str)
					p.typ()
				} else {
					p.sc.Mark("type name expected")
				}
			} else {
				p.sc.Mark("identifier expected")
			}
		case scanner.Pre:
			p.next()
		case scanner.Post:
			p.next()
		default:
			stop = true
		}
		p.pass(scanner.Separator, scanner.Delimiter)
	}
	if p.await(scanner.End, scanner.Delimiter, scanner.Separator) {
		p.next()
	} else {
		p.sc.Mark("END expected")
	}
}

func (p *pr) field() {
	fmt.Println("FIELD ", p.sym.Str)
	p.next()
	switch p.sym.Code {
	case scanner.Times:
		p.next()
		fmt.Println("exported")
	case scanner.Plus:
		p.next()
		fmt.Println("semi exported")
	case scanner.Minus:
		p.next()
		fmt.Println("semi hidden")
	}

	if p.await(scanner.Ident, scanner.Separator) {
		p.typ()
	} else if p.sym.Code == scanner.Proc {
		p.procTyp()
	} else {
		p.sc.Mark("unexpected field typ")
	}
}

func (p *pr) mapTyp() {
	assert.For(p.sym.Code == scanner.Ident, 20, "identifier expected")
	t, ok := entries[p.sym.Str].(Type)
	assert.For(ok && t.typ == ir.Map, 21, "map expected here")
	p.next()
	if p.await(scanner.Of, scanner.Separator) {
		//simple key:value map
		p.next()
		if p.await(scanner.Ident, scanner.Separator) {
			fmt.Println("KEY TYP")
			p.typ()
			if p.await(scanner.Comma, scanner.Separator) {
				p.next()
				if p.await(scanner.Ident, scanner.Separator) {
					fmt.Println("VALUE TYP")
					p.typ()
				} else {
					p.sc.Mark("identifier expected")
				}
			} else {
				p.sc.Mark("comma expected")
			}
		} else {
			p.sc.Mark("identifier expected")
		}
	} else if p.await(scanner.Ident, scanner.Separator, scanner.Delimiter) {
		//map with predefined fields
		for {
			p.field()
			if p.await(scanner.Delimiter, scanner.Separator) {
				if p.await(scanner.End, scanner.Separator, scanner.Delimiter) {
					break
				} else if p.sym.Code == scanner.Ident {
					//очередное поле
				} else {
					p.sc.Mark("identifier expected")
				}
			} else {
				p.sc.Mark("map delimiter expected")
			}
		}
		if p.await(scanner.End, scanner.Separator, scanner.Delimiter) {
			p.next()
		} else {
			p.sc.Mark("END expected")
		}
	} else {
		p.sc.Mark("incorrect map definition")
	}
}

func (p *pr) setTyp() {
	assert.For(p.sym.Code == scanner.Ident, 20, "identifier expected")
	t, ok := entries[p.sym.Str].(Type)
	assert.For(ok && t.typ == ir.Set, 21, "set expected here")
	p.next()
	if p.await(scanner.Of, scanner.Separator) {
		p.next()
		fmt.Println("OF")
		if p.await(scanner.Ident, scanner.Separator) {
			p.typ()
		}
		if p.await(scanner.With, scanner.Delimiter, scanner.Separator) {
			//expect allowed values
			fmt.Println("SET CONTENT SKIPPED")
			for p.next().Code != scanner.End {
			}
			p.next()
		}
	} else {
		p.sc.Mark("incorrect set definition")
	}
}

func (p *pr) typ() {
	assert.For(p.sym.Code == scanner.Ident, 20, "identifier expected")
	if t, ok := entries[p.sym.Str].(Type); ok {
		switch t.typ {
		case ir.Pointer:
			fmt.Println("POINTER")
			p.next()
			if p.await(scanner.To, scanner.Separator) {
				p.next()
				fmt.Println("TO")
				if p.await(scanner.Ident, scanner.Separator) {
					fmt.Println(p.sym.Str)
					if t0, ok := entries[p.sym.Str].(Type); ok {
						if t0.typ == ir.Map || t0.typ == ir.Set || t0.typ == ir.List {
							p.typ()
						}
					} else {
						p.next()
					}
				} else {
					p.sc.Mark("identifier expected")
				}
			} else {
				p.sc.Mark("TO expected")
			}
		case ir.Map:
			fmt.Println("MAP")
			p.mapTyp()
		case ir.String:
			fmt.Println("STRING")
			p.next()
		case ir.Integer:
			fmt.Println("INTEGER")
			p.next()
		case ir.Boolean:
			fmt.Println("BOOLEAN")
			p.next()
		case ir.Atom:
			fmt.Println("ATOM")
			p.next()
		case ir.List:
			fmt.Println("LIST")
			p.listTyp()
		case ir.Set:
			fmt.Println("SET")
			p.setTyp()
		default:
			p.sc.Mark("unexpected or unknown type", t.typ)
		}
	} else {
		fmt.Println("TYPE " + p.sym.Str)
		p.next()
	}
}

func (p *pr) typeDecl() {
	assert.For(p.sym.Code == scanner.Type, 20, "type section here")
	p.next()
	if p.sym.Code == scanner.Separator || p.sym.Code == scanner.Delimiter {
		for {
			if p.await(scanner.Ident, scanner.Separator, scanner.Delimiter) {
				fmt.Println("TYPE ", p.sym.Str)
				p.next()
				if p.sym.Code == scanner.Times {
					p.next()
					fmt.Println("exported")
				}
				if p.await(scanner.Ident, scanner.Separator) {
					p.typ()
					if p.await(scanner.Delimiter, scanner.Separator) {
						p.next()
					} else {
						p.sc.Mark("type delimiter expected")
					}
				} else {
					p.sc.Mark("base type expected")
				}
			} else {
				break
			}
		}
	} else {
		p.sc.Mark("separator expected")
	}
}

func (p *pr) varDecl() {
	assert.For(p.sym.Code == scanner.Var, 20, "var section here")
	p.next()
	if p.sym.Code == scanner.Separator || p.sym.Code == scanner.Delimiter {
		for {
			if p.await(scanner.Ident, scanner.Separator, scanner.Delimiter) {
				fmt.Println("VAR ", p.sym.Str)
				p.tg.BeginObject(target.Variable)
				p.tg.Name(p.sym.Str)
				p.next()
				switch p.sym.Code {
				case scanner.Times:
					p.next()
					fmt.Println("exported")
				case scanner.Plus:
					p.next()
					fmt.Println("semi exported")
				case scanner.Minus:
					p.next()
					fmt.Println("semi hidden")
				}
				if p.await(scanner.Ident, scanner.Separator) {
					p.typ()
					if p.await(scanner.Delimiter, scanner.Separator) {
						p.next()
					} else {
						p.sc.Mark("var delimiter expected")
					}
				} else {
					p.sc.Mark("var type expected")
				}
			} else {
				break
			}
		}
	} else {
		p.sc.Mark("separator expected")
	}
}

func (p *pr) statSeq() {
	p.pass(scanner.Delimiter, scanner.Separator)
	stop := false
	for !stop {
		switch p.sym.Code {
		case scanner.Close, scanner.End: //do nothing
			stop = true
		default:
			p.sc.Mark("unexpected ", p.sym)
		}
	}
}

func (p *pr) Module() (err error) {
	fmt.Println("COMPILER")
	if p.await(scanner.Module, scanner.Separator, scanner.Delimiter) {
		p.next()
		if p.await(scanner.Ident, scanner.Separator) {
			fmt.Println("MODULE", p.sym.Str)
			p.tg.Open(p.sym.Str)
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
				for p.await(scanner.Type, scanner.Delimiter, scanner.Separator) {
					p.typeDecl()
				}
				for p.await(scanner.Var, scanner.Delimiter, scanner.Separator) {
					p.varDecl()
				}
				if p.await(scanner.Begin, scanner.Separator, scanner.Delimiter) {
					p.next()
					p.statSeq()
				}
				if p.await(scanner.Close, scanner.Separator, scanner.Delimiter) {
					p.next()
					p.statSeq()
				}
				if p.await(scanner.End, scanner.Separator, scanner.Delimiter) {
					p.next()
					if p.await(scanner.Ident, scanner.Separator) {
						fmt.Println("END ", p.sym.Str)
						p.tg.Close(p.sym.Str)
						p.next()
						if p.await(scanner.Period) {
							p.next()
							fmt.Println("end compilation")
						} else {
							p.sc.Mark("period expected")
						}
					} else {
						p.sc.Mark("module name expected")
					}
				} else {
					p.sc.Mark("END expected")
				}
			} else {
				p.sc.Mark("mod delimiter expected")
			}
		} else {
			p.sc.Mark("module name expected")
		}
	} else {
		p.sc.Mark("MODULE expected")
	}
	return
}

func ConnectTo(s scanner.Scanner, t target.Target) Parser {
	assert.For(s != nil, 20)
	assert.For(t != nil, 21)
	ret := &pr{sc: s, tg: t}
	ret.init()
	return ret
}
