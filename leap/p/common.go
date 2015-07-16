package p

import (
	"fmt"
	"github.com/kpmy/ypk/assert"
	"github.com/kpmy/ypk/halt"
	"leaf/ir"
	"leaf/ir/operation"
	"leaf/ir/types"
	"leaf/leap/lss"
	"strconv"
)

type mark struct {
	rd        int
	line, col int
	marker    Marker
}

func (m *mark) Mark(msg ...interface{}) {
	m.marker.(*common).m = m
	m.marker.Mark(msg...)
}

func (m *mark) FutureMark() Marker { halt.As(100); return nil }

type common struct {
	sc    lss.Scanner
	sym   lss.Sym
	done  bool
	debug bool
	m     *mark
}

func (p *common) Mark(msg ...interface{}) {
	p.mark(msg...)
}

func (p *common) FutureMark() Marker {
	rd := p.sc.Read()
	str, pos := p.sc.Pos()
	m := &mark{marker: p, rd: rd, line: str, col: pos}
	return m
}

func (p *common) mark(msg ...interface{}) {
	rd := p.sc.Read()
	str, pos := p.sc.Pos()
	if len(msg) == 0 {
		p.m = &mark{rd: rd, line: str, col: pos}
	} else if p.m != nil {
		rd, str, pos = p.m.rd, p.m.line, p.m.col
		p.m = nil
	}
	if p.m == nil {
		panic(lss.Err("parser", rd, str, pos, msg...))
	}
}

func (p *common) next() lss.Sym {
	p.done = true
	if p.sym.Code != lss.None {
		//		fmt.Print("this ")
		//		fmt.Print("`" + fmt.Sprint(p.sym) + "`")
	}
	p.sym = p.sc.Get()
	//fmt.Print(" next ")
	if p.debug {
		fmt.Println("`" + fmt.Sprint(p.sym) + "`")
	}
	return p.sym
}

//expect is the most powerful step forward runner, breaks the compilation if unexpected sym found
func (p *common) expect(sym lss.Symbol, msg string, skip ...lss.Symbol) {
	assert.For(p.done, 20)
	if !p.await(sym, skip...) {
		p.mark(msg)
	}
	p.done = false
}

//await runs for the sym through skip list, but may not find the sym
func (p *common) await(sym lss.Symbol, skip ...lss.Symbol) bool {
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
func (p *common) pass(skip ...lss.Symbol) {
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
func (p *common) run(sym lss.Symbol) {
	if p.sym.Code != sym {
		for p.sc.Error() == nil && p.next().Code != sym {
			if p.sc.Error() != nil {
				p.mark("not found")
				break
			}
		}
	}
}

func (p *common) ident() string {
	assert.For(p.sym.Code == lss.Ident, 20, "identifier expected")
	//добавить валидацию идентификаторов
	return p.sym.Str
}

func (p *common) qualident(b *block) (id string, mod bool) {
	assert.For(p.is(lss.Ident), 20, "identifier expected here")
	id = p.ident()
	imp := b.imp(id)
	p.next()
	if p.is(lss.Period) && imp != nil {
		mod = true
		p.next()
		p.expect(lss.Ident, "identifier expected")
	}
	return
}

func (p *common) typ(consume func(t types.Type)) {
	assert.For(p.sym.Code == lss.Ident, 20, "type identifier expected here but found ", p.sym.Code)
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
		case types.ANY, types.PTR:
			p.next()
			consume(t.typ)
		case types.LIST, types.SET, types.MAP:
			p.next()
			consume(t.typ)
		default:
			p.mark("unexpected type ", id)
		}
	} else {
		p.mark("unknown type ", id)
	}
}

func (p *common) is(sym lss.Symbol) bool {
	return p.sym.Code == sym
}

func (p *common) number() (t types.Type, v interface{}) {
	assert.For(p.is(lss.Number), 20, "number expected here")
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

func (p *common) selector(b *selBuilder) {
	for stop := false; !stop; {
		p.pass(lss.Separator)
		switch p.sym.Code {
		case lss.Lbrak:
			p.next()
			this := &ir.SelectInside{}
			expr := &exprBuilder{sc: b.sc}
			p.expression(expr)
			this.Expr = expr
			b.join(this)
			p.expect(lss.Rbrak, "no ] found", lss.Separator)
			p.next()
		case lss.Period:
			p.next()
			this := &ir.SelectInside{}
			expr := &exprBuilder{sc: b.sc}
			p.factor(expr)
			this.Expr = expr
			b.join(this)
		case lss.Deref:
			p.next()
			this := &ir.SelectInside{}
			expr := &exprBuilder{sc: b.sc}
			expr.factor(&ir.ConstExpr{Type: types.Undef})
			this.Expr = expr
			b.join(this)
		default:
			stop = true
		}
	}
}

func (p *common) factor(b *exprBuilder) {
	b.marker = p
	switch p.sym.Code {
	case lss.String:
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
	case lss.Number:
		val := &ir.ConstExpr{}
		val.Type, val.Value = p.number()
		b.factor(val)
		p.next()
	case lss.True, lss.False:
		val := &ir.ConstExpr{}
		val.Type = types.BOOLEAN
		val.Value = (p.sym.Code == lss.True)
		b.factor(val)
		p.next()
	case lss.Null:
		val := &ir.ConstExpr{}
		val.Type = types.TRILEAN
		b.factor(val)
		p.next()
	case lss.Undef:
		val := &ir.ConstExpr{}
		val.Type = types.ANY
		b.factor(val)
		p.next()
	case lss.Nil:
		val := &ir.ConstExpr{}
		val.Type = types.PTR
		b.factor(val)
		p.next()
	case lss.Inf:
		val := &ir.ConstExpr{}
		val.Type = types.REAL
		val.Value = types.INF
		b.factor(val)
		p.next()
	case lss.Im:
		p.next()
		p.factor(b)
		p.pass(lss.Separator)
		b.factor(&ir.Monadic{Op: operation.Im})
	case lss.Not:
		p.next()
		p.factor(b)
		p.pass(lss.Separator)
		b.factor(&ir.Monadic{Op: operation.Not})
	case lss.Ident:
		mid, mod := p.qualident(b.sc)
		id := ""
		var before ir.Selector
		var e ir.Expression
		var after *selBuilder
		if mod {
			id = p.ident()
			e = b.asImp(mid, id)
			if _, ok := e.(*ir.NamedConstExpr); !ok {
				base := &selBuilder{sc: b.sc}
				imp := b.sc.im[mid]
				base.join(&ir.SelectMod{Mod: imp.Name})
				before = base
			}
			p.next()
		} else {
			id = mid
			mid = ""
			e = b.as(id)
		}
		after = &selBuilder{sc: b.sc}
		p.selector(after)
		b.factor(after.apply(before, e))
	case lss.Lparen:
		p.next()
		expr := &exprBuilder{sc: b.sc}
		p.expression(expr)
		b.factor(expr)
		p.expect(lss.Rparen, ") expected", lss.Separator)
		p.next()
	case lss.Lbrace:
		p.next()
		r := &ir.SetExpr{}
		for stop := false; !stop; {
			p.pass(lss.Separator, lss.Delimiter)
			if !p.is(lss.Rbrace) {
				expr := &exprBuilder{sc: b.sc}
				p.expression(expr)
				r.Expr = append(r.Expr, expr)
				if p.await(lss.Comma, lss.Separator) {
					p.next()
				} else {
					stop = true
				}
			} else { //empty set
				stop = true
			}
		}
		p.expect(lss.Rbrace, "} expected", lss.Separator)
		p.next()
		b.factor(r)
	case lss.Lbrak:
		p.next()
		r := &ir.ListExpr{}
		for stop := false; !stop; {
			p.pass(lss.Separator, lss.Delimiter)
			if !p.is(lss.Rbrak) {
				expr := &exprBuilder{sc: b.sc}
				p.expression(expr)
				r.Expr = append(r.Expr, expr)
				if p.await(lss.Comma, lss.Separator) {
					p.next()
				} else {
					stop = true
				}
			} else { //empty set
				stop = true
			}
		}
		p.expect(lss.Rbrak, "] expected", lss.Separator)
		p.next()
		b.factor(r)
	case lss.Lbrux:
		p.next()
		r := &ir.MapExpr{}
		for stop := false; !stop; {
			p.pass(lss.Separator, lss.Delimiter)
			if !p.is(lss.Rbrux) {
				kexpr := &exprBuilder{sc: b.sc}
				p.expression(kexpr)
				r.Key = append(r.Key, kexpr)
				p.expect(lss.Colon, "colon expected", lss.Separator)
				p.next()
				vexpr := &exprBuilder{sc: b.sc}
				p.expression(vexpr)
				r.Value = append(r.Value, vexpr)
				if p.await(lss.Comma, lss.Separator) {
					p.next()
				} else {
					stop = true
				}
			} else {
				stop = true
			}
		}
		p.expect(lss.Rbrux, "] expected", lss.Separator)
		p.next()
		b.factor(r)
	case lss.Infixate:
		p.next()
		p.expect(lss.Ident, "identifier expected")
		mid, mod := p.qualident(b.sc)
		id := ""
		if mod {
			id = p.ident()
			p.next()
		} else {
			id = mid
			mid = ""
		}
		limit := 0
		for stop := false; !stop; {
			p.expression(b)
			limit++
			if p.await(lss.Comma, lss.Separator) {
				p.next()
			} else {
				stop = true
			}
		}
		if limit > 1 {
			p.mark("expected one arg")
		}
		b.factor(b.infix(mid, id, limit))
	default:
		p.mark(p.sym, " is not an expression")
	}
}

func (p *common) cpx(b *exprBuilder) {
	b.marker = p
	p.factor(b)
	p.pass(lss.Separator)
	switch p.sym.Code {
	case lss.Ncmp, lss.Pcmp:
		op := p.sym.Code
		p.next()
		p.pass(lss.Separator)
		if p.sym.Code != lss.Im {
			p.factor(b)
		} else {
			p.mark("imaginary operator not expected")
		}
		b.power(&ir.Dyadic{Op: operation.Map(op)})

	}
}

func (p *common) power(b *exprBuilder) {
	b.marker = p
	p.cpx(b)
	for stop := false; !stop; {
		p.pass(lss.Separator)
		switch p.sym.Code {
		case lss.ArrowUp:
			op := p.sym.Code
			p.next()
			p.pass(lss.Separator)
			p.cpx(b)
			b.power(&ir.Dyadic{Op: operation.Map(op)})
		default:
			stop = true
		}
	}
}

func (p *common) product(b *exprBuilder) {
	b.marker = p
	p.power(b)
	for stop := false; !stop; {
		p.pass(lss.Separator)
		switch p.sym.Code {
		case lss.Times, lss.Div, lss.Mod, lss.Divide, lss.And:
			op := p.sym.Code
			p.next()
			p.pass(lss.Separator)
			p.power(b)
			b.product(&ir.Dyadic{Op: operation.Map(op)})
		default:
			stop = true
		}
	}
}

func (p *common) quantum(b *exprBuilder) {
	b.marker = p
	if p.is(lss.Minus) {
		p.next()
		p.pass(lss.Separator)
		p.product(b)
		b.product(&ir.Monadic{Op: operation.Neg})
	} else if p.is(lss.Plus) {
		p.next()
		p.pass(lss.Separator)
		p.product(b)
	} else {
		p.pass(lss.Separator)
		p.product(b)
	}
	for stop := false; !stop; {
		p.pass(lss.Separator)
		switch p.sym.Code {
		case lss.Plus, lss.Minus, lss.Or:
			op := p.sym.Code
			p.next()
			p.pass(lss.Separator)
			p.product(b)
			b.quantum(&ir.Dyadic{Op: operation.Map(op)})
		default:
			stop = true
		}
	}
}

func (p *common) expression(b *exprBuilder) {
	b.marker = p
	p.quantum(b)
	p.pass(lss.Separator)
	switch p.sym.Code {
	case lss.Equal, lss.Nequal, lss.Geq, lss.Leq, lss.Gtr, lss.Lss, lss.In:
		op := p.sym.Code
		p.next()
		p.pass(lss.Separator)
		p.quantum(b)
		b.expr(&ir.Dyadic{Op: operation.Map(op)})
	case lss.Infixate:
		p.next()
		p.expect(lss.Ident, "identifier expected")
		mid, mod := p.qualident(b.sc)
		id := ""
		if mod {
			id = p.ident()
			p.next()
		} else {
			id = mid
			mid = ""
		}
		limit := 1
		for stop := false; !stop; {
			p.quantum(b)
			limit++
			if p.await(lss.Comma, lss.Separator) {
				p.next()
			} else {
				stop = true
			}
		}
		if limit < 2 {
			p.mark("expected two or more args")
		}
		b.expr(b.infix(mid, id, limit))
	case lss.Is:
		p.next()
		p.pass(lss.Separator)
		p.typ(func(t types.Type) {
			b.expr(&ir.TypeTest{Typ: t})
		})
	}
}
