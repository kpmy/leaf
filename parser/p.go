package parser

import (
	"fmt"
	"github.com/kpmy/ypk/assert"
	"leaf/ir"
	"leaf/ir/types"
	"leaf/scanner"
	"strconv"
)

type Type struct {
	typ types.Type
}

var entries map[string]interface{}

func init() {
	entries = map[string]interface{}{"INTEGER": Type{typ: types.INTEGER}}
}

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
	target
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
	if p.sym.Code != sym {
		for p.next().Code != sym {
			if p.sc.Error() != nil {
				p.sc.Mark("not found")
				break
			}
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

func (p *pr) number() (types.Type, interface{}) {
	assert.For(p.is(scanner.Number), 20, "number expected here")
	x, err := strconv.Atoi(p.sym.Str)
	assert.For(err == nil, 40)
	return types.INTEGER, x
}

func (p *pr) factor(b *exprBuilder) {
	switch p.sym.Code {
	case scanner.Number:
		val := &ir.ConstExpr{}
		val.Type, val.Value = p.number()
		b.factor(val)
		p.next()
	case scanner.Ident:
		e := b.as(p.ident())
		b.factor(e)
		p.next()
	default:
		p.sc.Mark("not implemented for ", p.sym)
	}
}

func (p *pr) product(b *exprBuilder) {
	p.factor(b)
}

func (p *pr) quantum(b *exprBuilder) {
	p.product(b)
}

func (p *pr) expression(b *exprBuilder) {
	p.quantum(b)
}

func (p *pr) constDecl() {
	assert.For(p.sym.Code == scanner.Const, 20, "CONST block expected")
	p.next()
	for {
		if p.await(scanner.Ident, scanner.Delimiter, scanner.Separator) {
			id := p.ident()
			if p.root.ConstDecl[id] != nil {
				p.sc.Mark("identifier already exists")
			}
			p.next()
			obj := &ir.Const{Name: id}
			if p.await(scanner.Equal, scanner.Separator) { //const expression
				p.next()
				p.pass(scanner.Separator)
				obj.Expr = &exprBuilder{}
				p.expression(obj.Expr.(*exprBuilder))
			} else if p.is(scanner.Delimiter) { //ATOM
				obj.Expr = &ir.AtomExpr{Value: id}
				p.next()
			} else {
				p.sc.Mark("delimiter or expression expected")
			}
			p.root.ConstDecl[id] = obj
		} else {
			break
		}
	}
}

func (p *pr) typ(cons func(t types.Type)) {
	assert.For(p.sym.Code == scanner.Ident, 20, "type identifier expected here")
	id := p.ident()
	if t, ok := entries[id].(Type); ok {
		switch t.typ {
		case types.INTEGER:
			p.next()
			cons(t.typ)
		default:
			p.sc.Mark("unexpected type")
		}
	} else {
		p.sc.Mark("unknown type")
	}
}

func (p *pr) varDecl() {
	assert.For(p.sym.Code == scanner.Var, 20, "VAR block expected")
	p.next()
	for {
		if p.await(scanner.Ident, scanner.Delimiter, scanner.Separator) {
			obj := &ir.Variable{}
			id := p.ident()
			if p.root.ConstDecl[id] != nil {
				p.sc.Mark("identifier already exists")
			}
			obj.Name = id
			p.next()
			p.expect(scanner.Ident, "type identifier expected", scanner.Separator)
			p.typ(func(t types.Type) {
				obj.Type = t
			})
			p.root.VarDecl[obj.Name] = obj
		} else {
			break
		}
	}

}

func (p *pr) stmtSeq(b *blockBuilder) {
	for stop := false; !stop; {
		p.pass(scanner.Separator, scanner.Delimiter)
		switch p.sym.Code {
		case scanner.Ident:
			obj := b.obj(p.ident())
			p.next()
			p.pass(scanner.Separator)
			if p.is(scanner.Becomes) {
				stmt := &ir.AssignStmt{}
				p.next()
				p.pass(scanner.Separator)
				expr := &exprBuilder{}
				expr.scope = b.scope
				p.expression(expr)
				stmt.Object = obj
				stmt.Expr = expr
				b.put(stmt)
			} else {
				p.sc.Mark("illegal statement")
			}
		default:
			stop = true
		}
	}
}

func (p *pr) Module() (ret *ir.Module, err error) {
	p.expect(scanner.Module, "MODULE expected", scanner.Delimiter, scanner.Separator)
	p.next()
	p.expect(scanner.Ident, "module name expected", scanner.Separator)
	p.target.init(p.ident())
	p.next()
	p.pass(scanner.Separator, scanner.Delimiter)
	for p.await(scanner.Const, scanner.Delimiter, scanner.Separator) {
		p.constDecl()
	}
	for p.await(scanner.Var, scanner.Delimiter, scanner.Separator) {
		p.varDecl()
	}
	if p.await(scanner.Begin, scanner.Delimiter, scanner.Separator) {
		p.next()
		b := &blockBuilder{}
		b.scope = scopeLevel{varScope: p.root.VarDecl, constScope: p.root.ConstDecl}
		p.stmtSeq(b)
		p.root.BeginSeq = b.seq
	}
	if p.await(scanner.Close, scanner.Delimiter, scanner.Separator) {
		p.next()
		b := &blockBuilder{}
		b.scope = scopeLevel{varScope: p.root.VarDecl, constScope: p.root.ConstDecl}
		p.stmtSeq(b)
		p.root.CloseSeq = b.seq
	}
	//p.run(scanner.End)
	p.expect(scanner.End, "no END", scanner.Delimiter, scanner.Separator)
	p.next()
	p.expect(scanner.Ident, "module name expected", scanner.Separator)
	if p.ident() != p.root.Name {
		p.sc.Mark("module name does not match")
	}
	p.next()
	p.expect(scanner.Period, "end of module expected")
	ret = p.root
	return
}

func ConnectTo(s scanner.Scanner) Parser {
	assert.For(s != nil, 20)
	ret := &pr{sc: s}
	ret.init()
	return ret
}
