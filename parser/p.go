package parser

import (
	"fmt"
	"github.com/kpmy/ypk/assert"
	"leaf/ir"
	"leaf/ir/operation"
	"leaf/ir/types"
	"leaf/scanner"
	"strconv"
)

type Type struct {
	typ types.Type
}

var entries map[scanner.Foreign]interface{}
var idents map[string]scanner.Foreign

const (
	none scanner.Foreign = iota
	integer
	boolean
	trilean
	char
	str
	atom
	flo //for real, real is builtin go function O_o
)

func init() {
	idents = map[string]scanner.Foreign{"INTEGER": integer,
		"BOOLEAN": boolean,
		"TRILEAN": trilean,
		"CHAR":    char,
		"STRING":  str,
		"ATOM":    atom,
		"REAL":    flo}

	entries = map[scanner.Foreign]interface{}{integer: Type{typ: types.INTEGER},
		boolean: Type{typ: types.BOOLEAN},
		trilean: Type{typ: types.TRILEAN},
		char:    Type{typ: types.CHAR},
		str:     Type{typ: types.STRING},
		atom:    Type{typ: types.ATOM},
		flo:     Type{typ: types.REAL}}
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

func (p *pr) mark(msg ...interface{}) {
	str, pos := p.sc.Pos()
	panic(fmt.Sprint("parser: ", "at pos ", str, " ", pos, " ", fmt.Sprint(msg...)))
}

func (p *pr) next() scanner.Sym {
	p.done = true
	if p.sym.Code != scanner.Null {
		//		fmt.Print("this ")
		//		fmt.Print("`" + fmt.Sprint(p.sym) + "`")
	}
	p.sym = p.sc.Get()
	//fmt.Print(" next ")
	//	fmt.Println("`" + fmt.Sprint(p.sym) + "`")
	return p.sym
}

func (p *pr) init() {
	for k, v := range idents {
		p.sc.Register(v, k)
	}
	p.next()
}

//expect is the most powerful step forward runner, breaks the compilation if unexpected sym found
func (p *pr) expect(sym scanner.Symbol, msg string, skip ...scanner.Symbol) {
	assert.For(p.done, 20)
	if !p.await(sym, skip...) {
		p.mark(msg)
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

	for sym != p.sym.Code && skipped() && p.sc.Error() == nil {
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
	for skipped() && p.sc.Error() == nil {
		p.next()
	}
}

//run runs to the first sym through any other sym
func (p *pr) run(sym scanner.Symbol) {
	if p.sym.Code != sym {
		for p.sc.Error() == nil && p.next().Code != sym {
			if p.sc.Error() != nil {
				p.mark("not found")
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

func (p *pr) number() (t types.Type, v interface{}) {
	assert.For(p.is(scanner.Number), 20, "number expected here")
	switch p.sym.NumberOpts.Modifier {
	case "":
		if p.sym.NumberOpts.Period {
			t, v = types.REAL, p.sym.Str
		} else {
			//x, err := strconv.Atoi(p.sym.Str)
			//assert.For(err == nil, 40)
			t, v = types.INTEGER, p.sym.Str
		}
	case "U":
		if p.sym.NumberOpts.Period {
			p.mark("hex integer value expected")
		}
		//fmt.Println(p.sym)
		if r, err := strconv.ParseUint(p.sym.Str, 16, 64); err == nil {
			t = types.CHAR
			v = rune(r)
		} else {
			p.mark("error while reading integer")
		}
	default:
		p.mark("unknown number format `", p.sym.NumberOpts.Modifier, "`")
	}
	return
}

func (p *pr) selector(b *selBuilder) {
	for stop := false; !stop; {
		p.pass(scanner.Separator)
		switch p.sym.Code {
		case scanner.Lbrak:
			p.next()
			this := &ir.SelectIndex{}
			expr := &exprBuilder{scope: b.scope}
			p.expression(expr)
			this.Expr = expr
			b.join(this)
			p.expect(scanner.Rbrak, "no ] found", scanner.Separator)
			p.next()
		default:
			stop = true
		}
	}
}

func (p *pr) factor(b *exprBuilder) {
	switch p.sym.Code {
	case scanner.String:
		val := &ir.ConstExpr{}
		if len(p.sym.Str) == 1 && p.sym.StringOpts.Apos { //do it symbol
			val.Type = types.CHAR
			val.Value = rune(p.sym.Str[0])
			b.factor(val)
			p.next()
		} else { //do string later
			val.Type = types.STRING
			val.Value = p.sym.Str
			b.factor(val)
			p.next()
		}
	case scanner.Number:
		val := &ir.ConstExpr{}
		val.Type, val.Value = p.number()
		b.factor(val)
		p.next()
	case scanner.True, scanner.False:
		val := &ir.ConstExpr{}
		val.Type = types.BOOLEAN
		val.Value = (p.sym.Code == scanner.True)
		b.factor(val)
		p.next()
	case scanner.Nil:
		val := &ir.ConstExpr{}
		val.Type = types.TRILEAN
		b.factor(val)
		p.next()
	case scanner.Not:
		p.next()
		p.factor(b)
		p.pass(scanner.Separator)
		b.factor(&ir.Monadic{Op: operation.Not})
	case scanner.Ident:
		e := b.as(p.ident())
		p.next()
		sel := &selBuilder{scope: b.scope}
		p.selector(sel)
		b.factor(sel.appy(e))
	case scanner.Lparen:
		p.next()
		expr := &exprBuilder{scope: b.scope}
		p.expression(expr)
		b.factor(expr)
		p.expect(scanner.Rparen, ") expected", scanner.Separator)
		p.next()
	default:
		p.mark("not implemented for ", p.sym)
	}
}

func (p *pr) power(b *exprBuilder) {
	p.factor(b)
	for stop := false; !stop; {
		p.pass(scanner.Separator)
		switch p.sym.Code {
		case scanner.Arrow:
			op := p.sym.Code
			p.next()
			p.pass(scanner.Separator)
			p.factor(b)
			b.power(&ir.Dyadic{Op: operation.Map(op)})
		default:
			stop = true
		}
	}
}

func (p *pr) product(b *exprBuilder) {
	p.power(b)
	for stop := false; !stop; {
		p.pass(scanner.Separator)
		switch p.sym.Code {
		case scanner.Times, scanner.Div, scanner.Mod, scanner.Divide, scanner.And:
			op := p.sym.Code
			p.next()
			p.pass(scanner.Separator)
			p.power(b)
			b.product(&ir.Dyadic{Op: operation.Map(op)})
		default:
			stop = true
		}
	}
}

func (p *pr) quantum(b *exprBuilder) {
	if p.is(scanner.Minus) {
		p.next()
		p.pass(scanner.Separator)
		p.product(b)
		b.product(&ir.Monadic{Op: operation.Neg})
	} else if p.is(scanner.Plus) {
		p.next()
		p.pass(scanner.Separator)
		p.product(b)
	} else {
		p.pass(scanner.Separator)
		p.product(b)
	}
	for stop := false; !stop; {
		p.pass(scanner.Separator)
		switch p.sym.Code {
		case scanner.Plus, scanner.Minus, scanner.Or:
			op := p.sym.Code
			p.next()
			p.pass(scanner.Separator)
			p.product(b)
			b.quantum(&ir.Dyadic{Op: operation.Map(op)})
		default:
			stop = true
		}
	}
}

func (p *pr) expression(b *exprBuilder) {
	p.quantum(b)
	p.pass(scanner.Separator)
	switch p.sym.Code {
	case scanner.Equal, scanner.Nequal, scanner.Geq, scanner.Leq, scanner.Gtr, scanner.Lss:
		op := p.sym.Code
		p.next()
		p.pass(scanner.Separator)
		p.quantum(b)
		b.expr(&ir.Dyadic{Op: operation.Map(op)})
	}
}

func (p *pr) constDecl() {
	assert.For(p.sym.Code == scanner.Const, 20, "CONST block expected")
	p.next()
	for {
		if p.await(scanner.Ident, scanner.Delimiter, scanner.Separator) {
			id := p.ident()
			if p.root.ConstDecl[id] != nil {
				p.mark("identifier already exists")
			}
			p.next()
			obj := &ir.Const{Name: id}
			if p.await(scanner.Equal, scanner.Separator) { //const expression
				p.next()
				p.pass(scanner.Separator)
				obj.Expr = &exprBuilder{scope: scopeLevel{constScope: p.root.ConstDecl}}
				p.expression(obj.Expr.(*exprBuilder))
			} else if p.is(scanner.Delimiter) { //ATOM
				obj.Expr = &ir.AtomExpr{Value: id}
				p.next()
			} else {
				p.mark("delimiter or expression expected")
			}
			p.root.ConstDecl[id] = obj
		} else {
			break
		}
	}
}

func (p *pr) typ(consume func(t types.Type)) {
	assert.For(p.sym.Code == scanner.Ident, 20, "type identifier expected here")
	id := p.ident()
	if t, ok := entries[p.sym.User].(Type); ok {
		switch t.typ {
		case types.INTEGER, types.REAL, types.BOOLEAN, types.TRILEAN:
			p.next()
			consume(t.typ)
		case types.CHAR, types.STRING:
			p.next()
			consume(t.typ)
		case types.ATOM:
			p.next()
			consume(t.typ)
		default:
			p.mark("unexpected type ", id)
		}
	} else {
		p.mark("unknown type ", id)
	}
}

func (p *pr) varDecl() {
	assert.For(p.sym.Code == scanner.Var, 20, "VAR block expected")
	p.next()
	for {
		if p.await(scanner.Ident, scanner.Delimiter, scanner.Separator) {
			var vl []*ir.Variable
			for {
				obj := &ir.Variable{}
				id := p.ident()
				if p.root.ConstDecl[id] != nil {
					p.mark("identifier already exists")
				}
				obj.Name = id
				vl = append(vl, obj)
				p.root.VarDecl[obj.Name] = obj
				p.next()
				if p.await(scanner.Comma, scanner.Separator) {
					p.next()
					p.pass(scanner.Separator)
				} else {
					break
				}
			}
			p.expect(scanner.Ident, "type identifier expected", scanner.Separator)
			p.typ(func(t types.Type) {
				for _, obj := range vl {
					obj.Type = t
				}
			})
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
			sel := &selBuilder{scope: b.scope}
			p.selector(sel)
			sel.head(obj)
			if p.is(scanner.Becomes) {
				stmt := &ir.AssignStmt{}
				p.next()
				p.pass(scanner.Separator)
				expr := &exprBuilder{}
				expr.scope = b.scope
				p.expression(expr)
				stmt.Sel = sel
				stmt.Expr = expr
				b.put(stmt)
				p.expect(scanner.Delimiter, "delimiter expected", scanner.Separator)
			} else {
				p.mark("illegal statement")
			}
		case scanner.If:
			stmt := &ir.IfStmt{}
			for stop := false; !stop; {
				switch p.sym.Code {
				case scanner.If, scanner.Elsif:
					p.next()
					p.pass(scanner.Separator)
					expr := &exprBuilder{scope: b.scope}
					p.expression(expr)
					p.expect(scanner.Then, "THEN not found", scanner.Separator)
					p.next()
					st := &blockBuilder{scope: b.scope}
					p.stmtSeq(st)
					p.pass(scanner.Separator, scanner.Delimiter)
					br := &ir.IfBranch{}
					br.Expr = expr
					br.Seq = st.seq
					stmt.Cond = append(stmt.Cond, br)
				case scanner.Else:
					p.next()
					st := &blockBuilder{scope: b.scope}
					p.stmtSeq(st)
					p.pass(scanner.Separator, scanner.Delimiter)
					br := &ir.ElseBranch{}
					br.Seq = st.seq
					stmt.Else = br
				case scanner.End:
					p.next()
					stop = true
				default:
					p.mark("END or ELSIF expected, but ", p.sym)
				}
			}
			b.put(stmt)
		default:
			stop = true
		}
	}
}

func (p *pr) Module() (ret *ir.Module, err error) {
	if !p.await(scanner.Module, scanner.Delimiter, scanner.Separator) {
		if p.sc.Error() != nil {
			return nil, p.sc.Error()
		} else {
			p.mark("MODULE expected")
		}
	}
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
		p.mark("module name does not match")
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
