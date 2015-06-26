package leap

import (
	"fmt"
	"github.com/kpmy/ypk/assert"
	"leaf/ir"
	"leaf/ir/operation"
	"leaf/ir/types"
	"leaf/lss"
	"strconv"
)

type common struct {
	sc   lss.Scanner
	sym  lss.Sym
	done bool
}

func (p *common) mark(msg ...interface{}) {
	str, pos := p.sc.Pos()
	panic(fmt.Sprint("parser: ", "at pos ", str, " ", pos, " ", fmt.Sprint(msg...)))
}

func (p *common) next() lss.Sym {
	p.done = true
	if p.sym.Code != lss.Null {
		//		fmt.Print("this ")
		//		fmt.Print("`" + fmt.Sprint(p.sym) + "`")
	}
	p.sym = p.sc.Get()
	//fmt.Print(" next ")
	//fmt.Println("`" + fmt.Sprint(p.sym) + "`")
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
			this := &ir.SelectIndex{}
			expr := &exprBuilder{sc: b.sc}
			p.expression(expr)
			this.Expr = expr
			b.join(this)
			p.expect(lss.Rbrak, "no ] found", lss.Separator)
			p.next()
		default:
			stop = true
		}
	}
}

func (p *common) factor(b *exprBuilder) {
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
	case lss.Nil:
		val := &ir.ConstExpr{}
		val.Type = types.TRILEAN
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
		e := b.as(p.ident())
		p.next()
		sel := &selBuilder{sc: b.sc}
		p.selector(sel)
		b.factor(sel.appy(e))
	case lss.Lparen:
		p.next()
		expr := &exprBuilder{sc: b.sc}
		p.expression(expr)
		b.factor(expr)
		p.expect(lss.Rparen, ") expected", lss.Separator)
		p.next()
	default:
		p.mark("not implemented for ", p.sym)
	}
}

func (p *common) cpx(b *exprBuilder) {
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
	p.cpx(b)
	for stop := false; !stop; {
		p.pass(lss.Separator)
		switch p.sym.Code {
		case lss.Arrow:
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
	p.quantum(b)
	p.pass(lss.Separator)
	switch p.sym.Code {
	case lss.Equal, lss.Nequal, lss.Geq, lss.Leq, lss.Gtr, lss.Lss:
		op := p.sym.Code
		p.next()
		p.pass(lss.Separator)
		p.quantum(b)
		b.expr(&ir.Dyadic{Op: operation.Map(op)})
	case lss.Infixate:
		p.next()
		p.expect(lss.Ident, "identifier expected")
		id := p.ident()
		p.next()
		limit := 1
		for stop := false; !stop; {
			if !p.await(lss.Delimiter, lss.Separator) {
				p.quantum(b)
				limit++
			} else {
				stop = true
				p.next()
			}
		}
		if limit < 2 {
			p.mark("expected two or more args")
		}
		b.expr(b.infix(id, limit))
	}
}
