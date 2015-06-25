package leap

import (
	"fmt"
	"github.com/kpmy/ypk/assert"
	"leaf/ir"
	"leaf/ir/modifiers"
	"leaf/ir/operation"
	"leaf/ir/types"
	"leaf/leap/scanner"
	"strconv"
)

type Type struct {
	typ types.Type
}

var entries map[scanner.Foreign]interface{}
var idents map[string]scanner.Foreign
var mods map[scanner.Symbol]modifiers.Modifier

const (
	none scanner.Foreign = iota
	integer
	comp
	flo //for real, real is builtin go function O_o
	char
	str
	atom
	boolean
	trilean
)

func init() {
	idents = map[string]scanner.Foreign{"INTEGER": integer,
		"BOOLEAN": boolean,
		"TRILEAN": trilean,
		"CHAR":    char,
		"STRING":  str,
		"ATOM":    atom,
		"REAL":    flo,
		"COMPLEX": comp}

	entries = map[scanner.Foreign]interface{}{integer: Type{typ: types.INTEGER},
		boolean: Type{typ: types.BOOLEAN},
		trilean: Type{typ: types.TRILEAN},
		char:    Type{typ: types.CHAR},
		str:     Type{typ: types.STRING},
		atom:    Type{typ: types.ATOM},
		flo:     Type{typ: types.REAL},
		comp:    Type{typ: types.COMPLEX}}

	mods = map[scanner.Symbol]modifiers.Modifier{scanner.Minus: modifiers.Semi}
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
	//fmt.Println("`" + fmt.Sprint(p.sym) + "`")
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
			expr := &exprBuilder{sc: b.sc}
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
	case scanner.Im:
		p.next()
		p.factor(b)
		p.pass(scanner.Separator)
		b.factor(&ir.Monadic{Op: operation.Im})
	case scanner.Not:
		p.next()
		p.factor(b)
		p.pass(scanner.Separator)
		b.factor(&ir.Monadic{Op: operation.Not})
	case scanner.Ident:
		e := b.as(p.ident())
		p.next()
		sel := &selBuilder{sc: b.sc}
		p.selector(sel)
		b.factor(sel.appy(e))
	case scanner.Lparen:
		p.next()
		expr := &exprBuilder{sc: b.sc}
		p.expression(expr)
		b.factor(expr)
		p.expect(scanner.Rparen, ") expected", scanner.Separator)
		p.next()
	default:
		p.mark("not implemented for ", p.sym)
	}
}

func (p *pr) cpx(b *exprBuilder) {
	p.factor(b)
	p.pass(scanner.Separator)
	switch p.sym.Code {
	case scanner.Ncmp, scanner.Pcmp:
		op := p.sym.Code
		p.next()
		p.pass(scanner.Separator)
		if p.sym.Code != scanner.Im {
			p.factor(b)
		} else {
			p.mark("imaginary operator not expected")
		}
		b.power(&ir.Dyadic{Op: operation.Map(op)})

	}
}

func (p *pr) power(b *exprBuilder) {
	p.cpx(b)
	for stop := false; !stop; {
		p.pass(scanner.Separator)
		switch p.sym.Code {
		case scanner.Arrow:
			op := p.sym.Code
			p.next()
			p.pass(scanner.Separator)
			p.cpx(b)
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

func (p *pr) constDecl(b *constBuilder) {
	assert.For(p.sym.Code == scanner.Const, 20, "CONST block expected")
	p.next()
	for {
		if p.await(scanner.Ident, scanner.Delimiter, scanner.Separator) {
			id := p.ident()
			if c, free := b.this(id); c != nil || !free {
				p.mark("identifier already exists")
			}
			p.next()
			obj := &ir.Const{Name: id}
			if p.await(scanner.Equal, scanner.Separator) { //const expression
				p.next()
				p.pass(scanner.Separator)
				obj.Expr = &exprBuilder{sc: b.sc}
				p.expression(obj.Expr.(*exprBuilder))
			} else if p.is(scanner.Delimiter) { //ATOM
				obj.Expr = &ir.AtomExpr{Value: id}
				p.next()
			} else {
				p.mark("delimiter or expression expected")
			}
			b.decl(id, obj)
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
		case types.INTEGER, types.REAL, types.COMPLEX:
			p.next()
			consume(t.typ)
		case types.CHAR, types.STRING:
			p.next()
			consume(t.typ)
		case types.ATOM, types.BOOLEAN, types.TRILEAN:
			p.next()
			consume(t.typ)
		default:
			p.mark("unexpected type ", id)
		}
	} else {
		p.mark("unknown type ", id)
	}
}

func (p *pr) varDecl(b *varBuilder) {
	assert.For(p.sym.Code == scanner.Var, 20, "VAR block expected")
	p.next()
	for {
		if p.await(scanner.Ident, scanner.Delimiter, scanner.Separator) {
			var vl []*ir.Variable
			for {
				obj := &ir.Variable{}
				id := p.ident()
				if v, free := b.this(id); v != nil || !free {
					p.mark("identifier already exists")
				}
				obj.Name = id
				vl = append(vl, obj)
				b.decl(obj.Name, obj)
				p.next()
				if p.await(scanner.Minus) {
					obj.Modifier = mods[p.sym.Code]
					p.next()
				}
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
			if id := p.ident(); b.isObj(id) {
				obj := b.obj(id)
				p.next()
				p.pass(scanner.Separator)
				sel := &selBuilder{sc: b.sc}
				p.selector(sel)
				sel.head(obj)
				if p.is(scanner.Becomes) {
					stmt := &ir.AssignStmt{}
					p.next()
					p.pass(scanner.Separator)
					expr := &exprBuilder{sc: b.sc}
					p.expression(expr)
					stmt.Sel = sel
					stmt.Expr = expr
					b.put(stmt)
					//p.expect(scanner.Delimiter, "delimiter expected", scanner.Separator)
				} else {
					p.mark("illegal statement")
				}
			} else {
				p.next()
				var param []*forwardParam
				if p.await(scanner.Lparen, scanner.Separator, scanner.Delimiter) {
					p.next()
					for {
						p.expect(scanner.Ident, "identifier expected", scanner.Separator, scanner.Delimiter)
						par := &forwardParam{name: p.ident()}
						p.next()
						p.expect(scanner.Colon, "colon expected", scanner.Separator)
						p.next()
						e := &exprBuilder{sc: b.sc}
						p.expression(e)
						par.expr = e
						param = append(param, par)
						if p.await(scanner.Comma, scanner.Separator, scanner.Delimiter) {
							p.next()
						} else {
							break
						}
					}
					p.expect(scanner.Rparen, "no ) found", scanner.Separator, scanner.Delimiter)
					p.next()
				}
				stmt := b.call(id, param)
				b.put(stmt)
			}
		case scanner.If:
			stmt := &ir.IfStmt{}
			first := true
			for stop := false; !stop; {
				switch p.sym.Code {
				case scanner.If, scanner.Elsif:
					if p.is(scanner.If) && !first {
						p.mark("ELSIF expected")
					}
					first = false
					p.next()
					p.pass(scanner.Separator)
					expr := &exprBuilder{sc: b.sc}
					p.expression(expr)
					p.expect(scanner.Then, "THEN not found", scanner.Separator)
					p.next()
					st := &blockBuilder{sc: b.sc}
					p.stmtSeq(st)
					p.pass(scanner.Separator, scanner.Delimiter)
					br := &ir.ConditionBranch{}
					br.Expr = expr
					br.Seq = st.seq
					stmt.Cond = append(stmt.Cond, br)
				case scanner.Else:
					p.next()
					st := &blockBuilder{sc: b.sc}
					p.stmtSeq(st)
					p.pass(scanner.Separator, scanner.Delimiter)
					br := &ir.ElseBranch{}
					br.Seq = st.seq
					stmt.Else = br
				case scanner.End:
					p.next()
					stop = true
				default:
					p.mark("END or ELSE/ELSIF expected, but ", p.sym)
				}
			}
			b.put(stmt)
		case scanner.While:
			stmt := &ir.WhileStmt{}
			first := true
			for stop := false; !stop; {
				switch p.sym.Code {
				case scanner.While, scanner.Elsif:
					if p.is(scanner.While) && !first {
						p.mark("ELSIF expected")
					}
					first = false
					p.next()
					p.pass(scanner.Separator)
					expr := &exprBuilder{sc: b.sc}
					p.expression(expr)
					p.expect(scanner.Do, "DO not found", scanner.Separator)
					p.next()
					st := &blockBuilder{sc: b.sc}
					p.stmtSeq(st)
					p.pass(scanner.Separator, scanner.Delimiter)
					br := &ir.ConditionBranch{}
					br.Expr = expr
					br.Seq = st.seq
					stmt.Cond = append(stmt.Cond, br)
				case scanner.End:
					p.next()
					stop = true
				default:
					p.mark("END or ELSIF expected, but ", p.sym)
				}
			}
			b.put(stmt)
		case scanner.Repeat:
			p.next()
			stmt := &ir.RepeatStmt{}
			br := &ir.ConditionBranch{}
			st := &blockBuilder{sc: b.sc}
			p.pass(scanner.Separator, scanner.Delimiter)
			p.stmtSeq(st)
			p.expect(scanner.Until, "UNTIL expected", scanner.Separator, scanner.Delimiter)
			p.next()
			expr := &exprBuilder{sc: b.sc}
			p.expression(expr)
			p.expect(scanner.Delimiter, "delimiter expected", scanner.Separator)
			p.next()
			br.Expr = expr
			br.Seq = st.seq
			stmt.Cond = br
			b.put(stmt)
		case scanner.Choose:
			p.next()
			stmt := &ir.ChooseStmt{}
			if !p.await(scanner.Of, scanner.Separator, scanner.Delimiter) {
				expr := &exprBuilder{sc: b.sc}
				p.expression(expr)
				stmt.Expr = expr
				p.next()
				p.expect(scanner.Of, "OF expected", scanner.Separator, scanner.Delimiter)
			}
			first := true
			for stop := false; !stop; {
				switch p.sym.Code {
				case scanner.Of, scanner.Opt:
					if p.is(scanner.Of) && !first {
						p.mark("ELSIF expected")
					}
					first = false
					p.next()
					p.pass(scanner.Separator, scanner.Delimiter)
					var bunch []ir.Expression
					for {
						expr := &exprBuilder{sc: b.sc}
						p.expression(expr)
						bunch = append(bunch, expr)
						if p.await(scanner.Colon, scanner.Separator) {
							p.next()
							break
						} else if !p.is(scanner.Comma) {
							p.mark("comma expected")
						} else {
							p.next()
						}
					}
					st := &blockBuilder{sc: b.sc}
					p.stmtSeq(st)
					p.pass(scanner.Separator, scanner.Delimiter)
					for _, e := range bunch {
						br := &ir.ConditionBranch{}
						br.Expr = e
						br.Seq = st.seq
						stmt.Cond = append(stmt.Cond, br)
					}
				case scanner.Else:
					p.next()
					st := &blockBuilder{sc: b.sc}
					p.stmtSeq(st)
					p.pass(scanner.Separator, scanner.Delimiter)
					br := &ir.ElseBranch{}
					br.Seq = st.seq
					stmt.Else = br
				case scanner.End:
					stop = true
					p.next()
				default:
					p.mark("END expected")
				}
			}
			b.put(stmt)
		default:
			stop = true
		}
	}
}

func (p *pr) procDecl(b *blockBuilder) {
	assert.For(p.is(scanner.Proc), 20, "PROCEDURE expected here")
	ret := &ir.Procedure{}
	ret.Init()
	p.next()
	p.expect(scanner.Ident, "procedure name expected", scanner.Separator)
	ret.Name = p.ident()
	p.next()
	p.st.push()
	this := p.st.this()
	p.block(this)
	p.expect(scanner.Begin, "BEGIN expected", scanner.Separator, scanner.Delimiter)
	p.next()
	b.decl(ret.Name, ret)
	proc := &blockBuilder{sc: this}
	p.stmtSeq(proc)
	ret.Seq = proc.seq
	ret.ConstDecl = this.cm
	ret.VarDecl = this.vm
	ret.ProcDecl = this.pm
	p.expect(scanner.End, "no END", scanner.Delimiter, scanner.Separator)
	p.next()
	p.expect(scanner.Ident, "procedure name expected", scanner.Separator)
	if p.ident() != ret.Name {
		p.mark("procedure name does not match")
	}
	p.st.pop()
	p.next()
}

func (p *pr) block(bl *block) {
	for p.await(scanner.Const, scanner.Delimiter, scanner.Separator) {
		b := &constBuilder{sc: bl}
		p.constDecl(b)
	}
	for p.await(scanner.Var, scanner.Delimiter, scanner.Separator) {
		b := &varBuilder{sc: bl}
		p.varDecl(b)
	}
	for p.await(scanner.Proc, scanner.Delimiter, scanner.Separator) {
		b := &blockBuilder{sc: bl}
		p.procDecl(b)
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
	p.st.push()
	top := p.st.this()
	p.block(top)
	p.st.pop()
	p.top.ConstDecl = top.cm
	p.top.VarDecl = top.vm
	p.top.ProcDecl = top.pm
	if p.await(scanner.Begin, scanner.Delimiter, scanner.Separator) {
		p.next()
		b := &blockBuilder{sc: top}
		p.stmtSeq(b)
		p.top.BeginSeq = b.seq
	}
	if p.await(scanner.Close, scanner.Delimiter, scanner.Separator) {
		p.next()
		b := &blockBuilder{sc: top}
		p.stmtSeq(b)
		p.top.CloseSeq = b.seq
	}
	//p.run(scanner.End)
	p.expect(scanner.End, "no END", scanner.Delimiter, scanner.Separator)
	p.next()
	p.expect(scanner.Ident, "module name expected", scanner.Separator)
	if p.ident() != p.top.Name {
		p.mark("module name does not match")
	}
	p.next()
	p.expect(scanner.Period, "end of module expected")
	ret = p.top
	return
}

func ConnectTo(s scanner.Scanner) Parser {
	assert.For(s != nil, 20)
	ret := &pr{sc: s}
	ret.init()
	return ret
}
