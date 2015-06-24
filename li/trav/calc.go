package trav

import (
	"github.com/kpmy/trigo"
	"github.com/kpmy/ypk/assert"
	"github.com/kpmy/ypk/halt"
	"leaf/ir/operation"
	"leaf/ir/types"
	"math/big"
)

type df map[operation.Operation]func(*value, *value) *value
type tm map[types.Type]interface{}

var dyadic tm

func calcDyadic(left *value, op operation.Operation, right *value) (ret *value) {
	if ml, _ := dyadic[left.typ].(tm); ml != nil {
		if dml, _ := ml[right.typ].(df); dml != nil {
			if df := dml[op]; df != nil {
				ret = df(left, right)
			} else {
				halt.As(102, "operation not found ", op, " for ", left.typ, " ", right.typ)
			}
		} else {
			halt.As(101, "unexpected right ", right.typ, " for ", left.typ, " ", op)
		}
	} else {
		halt.As(100, "unexpected left ", left.typ, " ", right.typ, " for ", op)
	}
	return
}

func putDyadic(l, r types.Type, op operation.Operation, fn func(*value, *value) *value) {
	dr := dyadic[l]
	if dr == nil {
		dr = make(tm)
		dyadic[l] = dr
	}
	fm := dr.(tm)[r]
	if fm == nil {
		fm = make(df)
		dr.(tm)[r] = fm
	}
	fmm := fm.(df)
	assert.For(fmm[op] == nil, 40)
	fmm[op] = fn
}

func i_(fn func(*value, *value) *big.Int) func(*value, *value) *value {
	return func(l *value, r *value) (ret *value) {
		ret = &value{typ: types.INTEGER}
		ret.val = ThisInt(fn(l, r))
		return
	}
}

func r_(fn func(*value, *value) *big.Rat) func(*value, *value) *value {
	return func(l *value, r *value) (ret *value) {
		ret = &value{typ: types.REAL}
		ret.val = ThisRat(fn(l, r))
		return
	}
}

func r_r_(fn func(*big.Rat, *value) *big.Rat) func(*value, *value) *big.Rat {
	return func(l *value, r *value) *big.Rat {
		lr := l.toReal()
		return fn(lr, r)
	}
}

func r_r_r_(fn func(*big.Rat, *big.Rat) *big.Rat) func(*big.Rat, *value) *big.Rat {
	return func(lr *big.Rat, r *value) *big.Rat {
		rr := r.toReal()
		return fn(lr, rr)
	}
}

func r_ir_(fn func(*big.Rat, *value) *big.Rat) func(*value, *value) *big.Rat {
	return func(l *value, r *value) *big.Rat {
		if l.typ == types.REAL {
			lr := l.toReal()
			return fn(lr, r)
		} else if l.typ == types.INTEGER {
			li := l.toInt()
			lr := big.NewRat(0, 1)
			lr.SetInt(li)
			return fn(lr, r)
		} else {
			halt.As(100, "cannot convert")
		}
		panic(0)
	}
}

func r_ir_ir_(fn func(*big.Rat, *big.Rat) *big.Rat) func(*big.Rat, *value) *big.Rat {
	return func(lr *big.Rat, r *value) *big.Rat {
		if r.typ == types.REAL {
			rr := r.toReal()
			return fn(lr, rr)
		} else if r.typ == types.INTEGER {
			ri := r.toInt()
			rr := big.NewRat(0, 1)
			rr.SetInt(ri)
			return fn(lr, rr)
		} else {
			halt.As(100, "cannot convert")
		}
		panic(0)
	}
}

func c_(fn func(*value, *value) *Cmp) func(*value, *value) *value {
	return func(l *value, r *value) (ret *value) {
		ret = &value{typ: types.COMPLEX}
		ret.val = ThisCmp(fn(l, r))
		return
	}
}

func c_c_(fn func(*Cmp, *value) *Cmp) func(*value, *value) *Cmp {
	return func(l *value, r *value) *Cmp {
		lc := l.toCmp()
		return fn(lc, r)
	}
}

func c_c_c_(fn func(*Cmp, *Cmp) *Cmp) func(*Cmp, *value) *Cmp {
	return func(lc *Cmp, r *value) *Cmp {
		rc := r.toCmp()
		return fn(lc, rc)
	}
}

func c_ir_(fn func(*big.Rat, *value) *Cmp) func(*value, *value) *Cmp {
	return func(l *value, r *value) *Cmp {
		if l.typ == types.REAL {
			lr := l.toReal()
			return fn(lr, r)
		} else if l.typ == types.INTEGER {
			li := l.toInt()
			lr := big.NewRat(0, 1)
			lr.SetInt(li)
			return fn(lr, r)
		} else {
			halt.As(100, "cannot convert")
		}
		panic(0)
	}
}

func c_ir_ir_(fn func(*big.Rat, *big.Rat) *Cmp) func(*big.Rat, *value) *Cmp {
	return func(lr *big.Rat, r *value) *Cmp {
		if r.typ == types.REAL {
			rr := r.toReal()
			return fn(lr, rr)
		} else if r.typ == types.INTEGER {
			ri := r.toInt()
			rr := big.NewRat(0, 1)
			rr.SetInt(ri)
			return fn(lr, rr)
		} else {
			halt.As(100, "cannot convert")
		}
		panic(0)
	}
}

func i_i_(fn func(*big.Int, *value) *big.Int) func(*value, *value) *big.Int {
	return func(l *value, r *value) *big.Int {
		li := l.toInt()
		return fn(li, r)
	}
}

func i_i_i_(fn func(*big.Int, *big.Int) *big.Int) func(*big.Int, *value) *big.Int {
	return func(li *big.Int, r *value) *big.Int {
		ri := r.toInt()
		return fn(li, ri)
	}
}

func b_(fn func(*value, *value) bool) func(*value, *value) *value {
	return func(l *value, r *value) (ret *value) {
		ret = &value{typ: types.BOOLEAN}
		ret.val = fn(l, r)
		return
	}
}

func b_i_(fn func(*big.Int, *value) bool) func(*value, *value) bool {
	return func(l *value, r *value) bool {
		li := l.toInt()
		return fn(li, r)
	}
}

func b_i_i_(fn func(*big.Int, *big.Int) bool) func(*big.Int, *value) bool {
	return func(li *big.Int, r *value) bool {
		ri := r.toInt()
		return fn(li, ri)
	}
}

func b_r_(fn func(*big.Rat, *value) bool) func(*value, *value) bool {
	return func(l *value, r *value) bool {
		lr := l.toReal()
		return fn(lr, r)
	}
}

func b_r_r_(fn func(*big.Rat, *big.Rat) bool) func(*big.Rat, *value) bool {
	return func(lr *big.Rat, r *value) bool {
		rr := r.toReal()
		return fn(lr, rr)
	}
}

func b_c_(fn func(rune, *value) bool) func(*value, *value) bool {
	return func(l *value, r *value) bool {
		lc := l.toRune()
		return fn(lc, r)
	}
}

func b_c_c_(fn func(rune, rune) bool) func(rune, *value) bool {
	return func(lc rune, r *value) bool {
		rc := r.toRune()
		return fn(lc, rc)
	}
}

func b_b_(fn func(bool, *value) bool) func(*value, *value) bool {
	return func(l *value, r *value) bool {
		lb := l.toBool()
		return fn(lb, r)
	}
}

func b_b_b_(fn func(bool, bool) bool) func(bool, *value) bool {
	return func(lb bool, r *value) bool {
		rb := r.toBool()
		return fn(lb, rb)
	}
}

func b_t_(fn func(tri.Trit, *value) bool) func(*value, *value) bool {
	return func(l *value, r *value) bool {
		lt := l.toTril()
		return fn(lt, r)
	}
}

func b_t_t_(fn func(tri.Trit, tri.Trit) bool) func(tri.Trit, *value) bool {
	return func(lt tri.Trit, r *value) bool {
		rt := r.toTril()
		return fn(lt, rt)
	}
}

func b_a_(fn func(*Atom, *value) bool) func(*value, *value) bool {
	return func(l *value, r *value) bool {
		la := l.toAtom()
		return fn(la, r)
	}
}

func b_a_a_(fn func(*Atom, *Atom) bool) func(*Atom, *value) bool {
	return func(la *Atom, r *value) bool {
		ra := r.toAtom()
		return fn(la, ra)
	}
}

func b_a_t_(fn func(*Atom, tri.Trit) bool) func(*Atom, *value) bool {
	return func(la *Atom, r *value) bool {
		rt := r.toTril()
		return fn(la, rt)
	}
}

func b_t_a_(fn func(tri.Trit, *Atom) bool) func(tri.Trit, *value) bool {
	return func(lt tri.Trit, r *value) bool {
		ra := r.toAtom()
		return fn(lt, ra)
	}
}

func b_t_b_(fn func(tri.Trit, bool) bool) func(tri.Trit, *value) bool {
	return func(lt tri.Trit, r *value) bool {
		rb := r.toBool()
		return fn(lt, rb)
	}
}

func b_b_t_(fn func(bool, tri.Trit) bool) func(bool, *value) bool {
	return func(lb bool, r *value) bool {
		rt := r.toTril()
		return fn(lb, rt)
	}
}

func s_(fn func(*value, *value) string) func(*value, *value) *value {
	return func(l *value, r *value) (ret *value) {
		ret = &value{typ: types.STRING}
		ret.val = fn(l, r)
		return
	}
}

func s_s_(fn func(string, *value) string) func(*value, *value) string {
	return func(l *value, r *value) string {
		ls := l.toStr()
		return fn(ls, r)
	}
}

func s_c_(fn func(rune, *value) string) func(*value, *value) string {
	return func(l *value, r *value) string {
		lc := l.toRune()
		return fn(lc, r)
	}
}

func s_c_c_(fn func(rune, rune) string) func(rune, *value) string {
	return func(lc rune, r *value) string {
		rc := r.toRune()
		return fn(lc, rc)
	}
}

func s_s_s_(fn func(string, string) string) func(string, *value) string {
	return func(ls string, r *value) string {
		rs := r.toStr()
		return fn(ls, rs)
	}
}

func s_s_c_(fn func(string, rune) string) func(string, *value) string {
	return func(ls string, r *value) string {
		rc := r.toRune()
		return fn(ls, rc)
	}
}

func s_c_s_(fn func(rune, string) string) func(rune, *value) string {
	return func(lc rune, r *value) string {
		rs := r.toStr()
		return fn(lc, rs)
	}
}

func b_s_(fn func(string, *value) bool) func(*value, *value) bool {
	return func(l *value, r *value) bool {
		ls := l.toStr()
		return fn(ls, r)
	}
}

func b_s_s_(fn func(string, string) bool) func(string, *value) bool {
	return func(ls string, r *value) bool {
		rs := r.toStr()
		return fn(ls, rs)
	}
}

const (
	lss = -1
	eq  = 0
	gtr = 1
)

func dyINTEGER() {
	putDyadic(types.INTEGER, types.INTEGER, operation.Sum,
		i_(i_i_(i_i_i_(func(l *big.Int, r *big.Int) *big.Int {
			return l.Add(l, r)
		}))))
	putDyadic(types.INTEGER, types.INTEGER, operation.Diff,
		i_(i_i_(i_i_i_(func(l *big.Int, r *big.Int) *big.Int {
			return l.Sub(l, r)
		}))))
	putDyadic(types.INTEGER, types.INTEGER, operation.Prod,
		i_(i_i_(i_i_i_(func(l *big.Int, r *big.Int) *big.Int {
			return l.Mul(l, r)
		}))))
	putDyadic(types.INTEGER, types.INTEGER, operation.Div,
		i_(i_i_(i_i_i_(func(l *big.Int, r *big.Int) *big.Int {
			return l.Div(l, r)
		}))))
	putDyadic(types.INTEGER, types.INTEGER, operation.Mod,
		i_(i_i_(i_i_i_(func(l *big.Int, r *big.Int) *big.Int {
			return l.Mod(l, r)
		}))))
	putDyadic(types.INTEGER, types.INTEGER, operation.Pow,
		i_(i_i_(i_i_i_(func(l *big.Int, r *big.Int) *big.Int {
			assert.For(r.Cmp(big.NewInt(0)) >= 0, 40, "nonnegative only")
			return l.Exp(l, r, big.NewInt(0))
		}))))

	putDyadic(types.INTEGER, types.INTEGER, operation.Lss,
		b_(b_i_(b_i_i_(func(l *big.Int, r *big.Int) bool {
			res := l.Cmp(r)
			return res == lss
		}))))
	putDyadic(types.INTEGER, types.INTEGER, operation.Gtr,
		b_(b_i_(b_i_i_(func(l *big.Int, r *big.Int) bool {
			res := l.Cmp(r)
			return res == gtr
		}))))
	putDyadic(types.INTEGER, types.INTEGER, operation.Leq,
		b_(b_i_(b_i_i_(func(l *big.Int, r *big.Int) bool {
			res := l.Cmp(r)
			return res != gtr
		}))))
	putDyadic(types.INTEGER, types.INTEGER, operation.Geq,
		b_(b_i_(b_i_i_(func(l *big.Int, r *big.Int) bool {
			res := l.Cmp(r)
			return res != lss
		}))))
	putDyadic(types.INTEGER, types.INTEGER, operation.Eq,
		b_(b_i_(b_i_i_(func(l *big.Int, r *big.Int) bool {
			res := l.Cmp(r)
			return res == eq
		}))))
	putDyadic(types.INTEGER, types.INTEGER, operation.Neq,
		b_(b_i_(b_i_i_(func(l *big.Int, r *big.Int) bool {
			res := l.Cmp(r)
			return res != eq
		}))))
}

func dyREAL() {
	putDyadic(types.REAL, types.REAL, operation.Sum,
		r_(r_r_(r_r_r_(func(l *big.Rat, r *big.Rat) *big.Rat {
			return l.Add(l, r)
		}))))
	putDyadic(types.REAL, types.REAL, operation.Diff,
		r_(r_r_(r_r_r_(func(l *big.Rat, r *big.Rat) *big.Rat {
			return l.Sub(l, r)
		}))))
	putDyadic(types.REAL, types.REAL, operation.Prod,
		r_(r_r_(r_r_r_(func(l *big.Rat, r *big.Rat) *big.Rat {
			return l.Mul(l, r)
		}))))
	putDyadic(types.REAL, types.REAL, operation.Quot,
		r_(r_r_(r_r_r_(func(l *big.Rat, r *big.Rat) *big.Rat {
			return l.Quo(l, r)
		}))))

	putDyadic(types.REAL, types.REAL, operation.Lss,
		b_(b_r_(b_r_r_(func(l *big.Rat, r *big.Rat) bool {
			res := l.Cmp(r)
			return res == lss
		}))))
	putDyadic(types.REAL, types.REAL, operation.Gtr,
		b_(b_r_(b_r_r_(func(l *big.Rat, r *big.Rat) bool {
			res := l.Cmp(r)
			return res == gtr
		}))))
	putDyadic(types.REAL, types.REAL, operation.Leq,
		b_(b_r_(b_r_r_(func(l *big.Rat, r *big.Rat) bool {
			res := l.Cmp(r)
			return res != gtr
		}))))
	putDyadic(types.REAL, types.REAL, operation.Geq,
		b_(b_r_(b_r_r_(func(l *big.Rat, r *big.Rat) bool {
			res := l.Cmp(r)
			return res != lss
		}))))
	putDyadic(types.REAL, types.REAL, operation.Eq,
		b_(b_r_(b_r_r_(func(l *big.Rat, r *big.Rat) bool {
			res := l.Cmp(r)
			return res == eq
		}))))
	putDyadic(types.REAL, types.REAL, operation.Neq,
		b_(b_r_(b_r_r_(func(l *big.Rat, r *big.Rat) bool {
			res := l.Cmp(r)
			return res != eq
		}))))
	putDyadic(types.REAL, types.REAL, operation.Pow,
		r_(r_r_(r_r_r_(func(l *big.Rat, r *big.Rat) *big.Rat {
			n := l.Num()
			d := l.Denom()
			p := r.Num()
			q := r.Denom()
			assert.For(p.Cmp(big.NewInt(0)) >= 0, 40, "nonnegative only")
			assert.For(q.Cmp(big.NewInt(1)) == 0, 41, "извлечение корня не поддерживается")
			n = n.Exp(n, p, nil)
			d = d.Exp(d, p, nil)
			ret := big.NewRat(0, 1)
			ret = ret.SetFrac(n, d)
			return ret
		}))))
}

func dyCOMPLEX() {
	putDyadic(types.COMPLEX, types.COMPLEX, operation.Sum,
		c_(c_c_(c_c_c_(func(l *Cmp, r *Cmp) *Cmp {
			ret := &Cmp{}
			ret.re = l.re.Add(l.re, r.re)
			ret.im = l.im.Add(l.im, r.im)
			return ret
		}))))
}

func dyCHAR() {
	putDyadic(types.CHAR, types.CHAR, operation.Eq, b_(b_c_(b_c_c_(func(lc rune, rc rune) bool { return lc == rc }))))
	putDyadic(types.CHAR, types.CHAR, operation.Neq, b_(b_c_(b_c_c_(func(lc rune, rc rune) bool { return lc != rc }))))
	putDyadic(types.CHAR, types.CHAR, operation.Leq, b_(b_c_(b_c_c_(func(lc rune, rc rune) bool { return lc <= rc }))))
	putDyadic(types.CHAR, types.CHAR, operation.Geq, b_(b_c_(b_c_c_(func(lc rune, rc rune) bool { return lc >= rc }))))
	putDyadic(types.CHAR, types.CHAR, operation.Lss, b_(b_c_(b_c_c_(func(lc rune, rc rune) bool { return lc < rc }))))
	putDyadic(types.CHAR, types.CHAR, operation.Gtr, b_(b_c_(b_c_c_(func(lc rune, rc rune) bool { return lc > rc }))))
}

func dySTRING() {
	putDyadic(types.STRING, types.STRING, operation.Sum, s_(s_s_(s_s_s_(func(ls string, rs string) string { return ls + rs }))))

	putDyadic(types.STRING, types.STRING, operation.Eq, b_(b_s_(b_s_s_(func(ls string, rs string) bool { return ls == rs }))))
	putDyadic(types.STRING, types.STRING, operation.Leq, b_(b_s_(b_s_s_(func(ls string, rs string) bool { return ls <= rs }))))
	putDyadic(types.STRING, types.STRING, operation.Lss, b_(b_s_(b_s_s_(func(ls string, rs string) bool { return ls < rs }))))
	putDyadic(types.STRING, types.STRING, operation.Geq, b_(b_s_(b_s_s_(func(ls string, rs string) bool { return ls >= rs }))))
	putDyadic(types.STRING, types.STRING, operation.Gtr, b_(b_s_(b_s_s_(func(ls string, rs string) bool { return ls > rs }))))
	putDyadic(types.STRING, types.STRING, operation.Neq, b_(b_s_(b_s_s_(func(ls string, rs string) bool { return ls != rs }))))
}

func dyINT2REAL() {
	putDyadic(types.REAL, types.INTEGER, operation.Quot,
		r_(r_ir_(r_ir_ir_(func(l *big.Rat, r *big.Rat) *big.Rat {
			return l.Quo(l, r)
		}))))
	putDyadic(types.INTEGER, types.REAL, operation.Quot,
		r_(r_ir_(r_ir_ir_(func(l *big.Rat, r *big.Rat) *big.Rat {
			return l.Quo(l, r)
		}))))
	putDyadic(types.REAL, types.INTEGER, operation.Pow,
		r_(r_ir_(r_ir_ir_(func(l *big.Rat, r *big.Rat) *big.Rat {
			n := l.Num()
			d := l.Denom()
			assert.For(r.IsInt(), 40)
			p := r.Num()
			assert.For(p.Cmp(big.NewInt(0)) >= 0, 40, "positive only")
			n = n.Exp(n, p, nil)
			d = d.Exp(d, p, nil)
			ret := big.NewRat(0, 1)
			ret = ret.SetFrac(n, d)
			return ret
		}))))
	putDyadic(types.INTEGER, types.REAL, operation.Pow,
		r_(r_ir_(r_ir_ir_(func(l *big.Rat, r *big.Rat) *big.Rat {
			assert.For(l.IsInt(), 40)
			n := l.Num()
			p := r.Num()
			q := r.Denom()
			assert.For(p.Cmp(big.NewInt(0)) >= 0, 40, "positive only")
			assert.For(q.Cmp(big.NewInt(1)) == 0, 41, "извлечение корня не поддерживается")
			n = n.Exp(n, p, q)
			ret := big.NewRat(0, 1)
			ret = ret.SetFrac(n, big.NewInt(1))
			return ret
		}))))
	putDyadic(types.INTEGER, types.REAL, operation.Prod,
		r_(r_ir_(r_ir_ir_(func(l *big.Rat, r *big.Rat) *big.Rat {
			return l.Mul(l, r)
		}))))
	putDyadic(types.REAL, types.INTEGER, operation.Prod,
		r_(r_ir_(r_ir_ir_(func(l *big.Rat, r *big.Rat) *big.Rat {
			return l.Mul(l, r)
		}))))
	putDyadic(types.REAL, types.INTEGER, operation.Sum,
		r_(r_ir_(r_ir_ir_(func(l *big.Rat, r *big.Rat) *big.Rat {
			return l.Add(l, r)
		}))))
	putDyadic(types.REAL, types.INTEGER, operation.Diff,
		r_(r_ir_(r_ir_ir_(func(l *big.Rat, r *big.Rat) *big.Rat {
			return l.Sub(l, r)
		}))))
	putDyadic(types.INTEGER, types.INTEGER, operation.Quot,
		r_(r_ir_(r_ir_ir_(func(l *big.Rat, r *big.Rat) *big.Rat {
			return l.Quo(l, r)
		}))))
}

func dyREAL2COMPLEX() {
	putDyadic(types.INTEGER, types.INTEGER, operation.Pcmp,
		c_(c_ir_(c_ir_ir_(func(l *big.Rat, r *big.Rat) *Cmp {
			return &Cmp{re: l, im: r}
		}))))
	putDyadic(types.INTEGER, types.INTEGER, operation.Ncmp,
		c_(c_ir_(c_ir_ir_(func(l *big.Rat, r *big.Rat) *Cmp {
			return &Cmp{re: l, im: r.Neg(r)}
		}))))
}

func dyCHAR2STRING() {
	putDyadic(types.STRING, types.CHAR, operation.Sum, s_(s_s_(s_s_c_(func(ls string, rc rune) string {
		buf := []rune(ls)
		buf = append(buf, rc)
		return string(buf)
	}))))

	putDyadic(types.CHAR, types.STRING, operation.Sum, s_(s_c_(s_c_s_(func(lc rune, rs string) string {
		var buf []rune
		buf = append(buf, lc)
		buf2 := []rune(rs)
		buf = append(buf, buf2...)
		return string(buf)
	}))))

	putDyadic(types.CHAR, types.CHAR, operation.Sum, s_(s_c_(s_c_c_(func(lc rune, rc rune) string {
		buf := []rune{lc, rc}
		return string(buf)
	}))))
}

func dyABT() {
	putDyadic(types.BOOLEAN, types.BOOLEAN, operation.Neq, b_(b_b_(b_b_b_(func(lb bool, rb bool) bool { return lb != rb }))))
	putDyadic(types.BOOLEAN, types.BOOLEAN, operation.Eq, b_(b_b_(b_b_b_(func(lb bool, rb bool) bool { return lb == rb }))))

	putDyadic(types.TRILEAN, types.TRILEAN, operation.Neq, b_(b_t_(b_t_t_(func(lt tri.Trit, rt tri.Trit) bool { return tri.Ord(lt) != tri.Ord(rt) }))))
	putDyadic(types.TRILEAN, types.TRILEAN, operation.Eq, b_(b_t_(b_t_t_(func(lt tri.Trit, rt tri.Trit) bool { return tri.Ord(lt) == tri.Ord(rt) }))))

	putDyadic(types.TRILEAN, types.BOOLEAN, operation.Eq, b_(b_t_(b_t_b_(func(lt tri.Trit, rb bool) bool {
		if !tri.Nil(lt) {
			if tri.True(lt) {
				rb = (rb == true)
			} else {
				rb = (rb == false)
			}
		} else {
			rb = false
		}
		return rb
	}))))

	putDyadic(types.TRILEAN, types.BOOLEAN, operation.Neq, b_(b_t_(b_t_b_(func(lt tri.Trit, rb bool) bool {
		if !tri.Nil(lt) {
			if tri.True(lt) {
				rb = (rb != true)
			} else {
				rb = (rb != false)
			}
		} else {
			rb = true
		}
		return rb
	}))))

	putDyadic(types.BOOLEAN, types.TRILEAN, operation.Neq, b_(b_b_(b_b_t_(func(lb bool, rt tri.Trit) bool {
		if !tri.Nil(rt) {
			if tri.True(rt) {
				lb = (lb != true)
			} else {
				lb = (lb != false)
			}
		} else {
			lb = true
		}
		return lb
	}))))

	putDyadic(types.BOOLEAN, types.TRILEAN, operation.Eq, b_(b_b_(b_b_t_(func(lb bool, rt tri.Trit) bool {
		if !tri.Nil(rt) {
			if tri.True(rt) {
				lb = (lb == true)
			} else {
				lb = (lb == false)
			}
		} else {
			lb = false
		}
		return lb
	}))))

	putDyadic(types.ATOM, types.ATOM, operation.Neq, b_(b_a_(b_a_a_(func(la *Atom, ra *Atom) bool {
		neq := true
		if la == nil && ra == nil {
			neq = false
		} else if la != nil && ra != nil {
			neq = *la != *ra
		}
		return neq
	}))))

	putDyadic(types.ATOM, types.ATOM, operation.Eq, b_(b_a_(b_a_a_(func(la *Atom, ra *Atom) bool {
		eq := false
		if la == nil && ra == nil {
			eq = true
		} else if la != nil && ra != nil {
			eq = *la == *ra
		}
		return eq
	}))))

	putDyadic(types.ATOM, types.TRILEAN, operation.Eq, b_(b_a_(b_a_t_(func(la *Atom, rt tri.Trit) bool {
		assert.For(tri.Nil(rt), 40, "NIL comparision only")
		return la == nil
	}))))

	putDyadic(types.TRILEAN, types.ATOM, operation.Eq, b_(b_t_(b_t_a_(func(lt tri.Trit, ra *Atom) bool {
		assert.For(tri.Nil(lt), 40, "NIL comparision only")
		return ra == nil
	}))))

	putDyadic(types.ATOM, types.TRILEAN, operation.Neq, b_(b_a_(b_a_t_(func(la *Atom, rt tri.Trit) bool {
		assert.For(tri.Nil(rt), 40, "NIL comparision only")
		return la != nil
	}))))

	putDyadic(types.TRILEAN, types.ATOM, operation.Neq, b_(b_t_(b_t_a_(func(lt tri.Trit, ra *Atom) bool {
		assert.For(tri.Nil(lt), 40, "NIL comparision only")
		return ra != nil
	}))))
}

func init() {
	dyadic = make(tm)
	dyINTEGER()
	dyREAL()
	dyCOMPLEX()
	dyCHAR()
	dySTRING()
	dyINT2REAL()
	dyREAL2COMPLEX()
	dyCHAR2STRING()
	dyABT()
}
