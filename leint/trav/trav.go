package trav

import (
	"container/list"
	"fmt"
	"github.com/kpmy/trigo"
	"github.com/kpmy/ypk/assert"
	"github.com/kpmy/ypk/halt"
	"leaf/ir"
	"leaf/ir/operation"
	"leaf/ir/types"
	"leaf/leint"
	"math/big"
	"reflect"
)

type Atom string

type Int struct {
	big.Int
}

func NewInt(x int64) (ret *Int) {
	ret = &Int{}
	ret.Int = *big.NewInt(x)
	return
}

func ThisInt(x *big.Int) (ret *Int) {
	ret = &Int{}
	ret.Int = *x
	return
}

func (i *Int) String() string {
	x, _ := i.Int.MarshalText()
	return string(x)
}

type Rat struct {
	big.Rat
}

func NewRat(x float64) (ret *Rat) {
	ret = &Rat{}
	ret.Rat = *big.NewRat(0, 1)
	return
}

func ThisRat(x *big.Rat) (ret *Rat) {
	ret = &Rat{}
	ret.Rat = *x
	return
}

type Cmp struct {
	re, im *big.Rat
}

func (c *Cmp) String() (ret string) {
	null := big.NewRat(0, 1)
	if c.re.Cmp(null) != 0 {
		ret = fmt.Sprint(c.re)
	}
	if eq := c.im.Cmp(null); eq > 0 {
		ret = fmt.Sprint(ret, "+i", c.im.Abs(c.im))
	} else if eq < 0 {
		ret = fmt.Sprint(ret, "-i", c.im.Abs(c.im))
	} else if ret == "" {
		ret = "0"
	}
	return
}

func NewCmp(re, im float64) (ret *Cmp) {
	ret = &Cmp{}
	ret.re = big.NewRat(0, 1).SetFloat64(re)
	ret.im = big.NewRat(0, 1).SetFloat64(im)
	return
}

func ThisCmp(c *Cmp) (ret *Cmp) {
	ret = &Cmp{}
	*ret = *c
	return
}

type storage struct {
	schema map[string]*ir.Variable
	data   map[string]interface{}
}

type stack struct {
	vl *list.List
}

type value struct {
	typ types.Type
	val interface{}
}

type context struct {
	storage
	stack
	root *ir.Module
}

func (v *value) toInt() (ret *big.Int) {
	assert.For(v.typ == types.INTEGER, 20)
	switch x := v.val.(type) {
	case int:
		ret = big.NewInt(int64(x))
	case *Int:
		ret = big.NewInt(0)
		ret.Add(ret, &x.Int)
	default:
		halt.As(100, "wrong integer ", reflect.TypeOf(x))
	}
	return
}

func (v *value) toReal() (ret *big.Rat) {
	assert.For(v.typ == types.REAL, 20)
	switch x := v.val.(type) {
	case *Rat:
		ret = big.NewRat(0, 1)
		ret.Add(ret, &x.Rat)
	default:
		halt.As(100, "wrong real ", reflect.TypeOf(x))
	}
	return
}

func (v *value) toBool() (ret bool) {
	assert.For(v.typ == types.BOOLEAN, 20)
	switch x := v.val.(type) {
	case bool:
		ret = x
	default:
		halt.As(100, "wrong boolean ", reflect.TypeOf(x))
	}
	return
}

func (v *value) toStr() (ret string) {
	assert.For(v.typ == types.STRING, 20)
	switch x := v.val.(type) {
	case string:
		ret = x
	default:
		halt.As(100, "wrong string ", reflect.TypeOf(x))
	}
	return
}

func (v *value) toRune() (ret rune) {
	assert.For(v.typ == types.CHAR, 20)
	switch x := v.val.(type) {
	case rune:
		ret = x
	default:
		halt.As(100, "wrong rune ", reflect.TypeOf(x))
	}
	return
}

func (v *value) toTril() (ret tri.Trit) {
	assert.For(v.typ == types.TRILEAN, 20)
	switch x := v.val.(type) {
	case tri.Trit:
		ret = x
	default:
		halt.As(100, "wrong trilean ", reflect.TypeOf(x))
	}
	return
}

func (v *value) toAtom() (ret *Atom) {
	assert.For(v.typ == types.ATOM, 20)
	switch x := v.val.(type) {
	case Atom:
		ret = &x
	case nil: //do nothing
	default:
		halt.As(100, "wrong atom ", reflect.TypeOf(x))
	}
	return
}

func (v *value) toCmp() (ret *Cmp) {
	assert.For(v.typ == types.COMPLEX, 20)
	switch x := v.val.(type) {
	case *Cmp:
		ret = ThisCmp(x)
	default:
		halt.As(100, "wrong complex ", reflect.TypeOf(x))
	}
	return
}

func cval(e *ir.ConstExpr) (ret *value) {
	t := e.Type
	switch t {
	case types.INTEGER:
		b := big.NewInt(0)
		if err := b.UnmarshalText([]byte(e.Value.(string))); err == nil {
			v := ThisInt(b)
			ret = &value{typ: t, val: v}
		} else {
			halt.As(100, "wrong integer")
		}
	case types.BOOLEAN:
		v := e.Value.(bool)
		ret = &value{typ: t, val: v}
	case types.CHAR:
		var v rune
		switch x := e.Value.(type) {
		case int32:
			v = rune(x)
		case int:
			v = rune(x)
		default:
			halt.As(100, "unsupported rune coding")
		}
		ret = &value{typ: t, val: v}
	case types.STRING:
		v := e.Value.(string)
		ret = &value{typ: t, val: v}
	case types.TRILEAN:
		if e.Value == nil {
			ret = &value{typ: t, val: tri.NIL}
		} else if tv := e.Value.(bool); tv {
			ret = &value{typ: t, val: tri.TRUE}
		} else {
			ret = &value{typ: t, val: tri.FALSE}
		}
	case types.REAL:
		r := big.NewRat(0, 1)
		if err := r.UnmarshalText([]byte(e.Value.(string))); err == nil {
			v := ThisRat(r)
			ret = &value{typ: t, val: v}
		} else {
			halt.As(100, "wrong real")
		}
	default:
		halt.As(100, "unknown type ", t, " for ", e)
	}
	return
}

func (s *storage) init() {
	s.schema = make(map[string]*ir.Variable)
	s.data = make(map[string]interface{})
}

func (s *storage) alloc(vl map[string]*ir.Variable) {
	if vl != nil {
		s.schema = vl
	}
	for _, v := range s.schema {
		switch v.Type {
		case types.INTEGER:
			s.data[v.Name] = NewInt(0)
		case types.BOOLEAN:
			s.data[v.Name] = false
		case types.TRILEAN:
			s.data[v.Name] = tri.NIL
		case types.CHAR:
			s.data[v.Name] = rune(0)
		case types.STRING:
			s.data[v.Name] = ""
		case types.ATOM:
			s.data[v.Name] = nil
		case types.REAL:
			s.data[v.Name] = NewRat(0.0)
		case types.COMPLEX:
			s.data[v.Name] = NewCmp(0.0, 0.0)
		default:
			halt.As(100, "unknown type ", v.Name, ": ", v.Type)
		}
	}
}

func (s *storage) findObj(o *ir.Variable, fn func(*value) *value) {
	data := s.data[o.Name]
	nv := fn(&value{typ: o.Type, val: data})
	if nv != nil {
		assert.For(nv.typ == o.Type, 40, "provided ", nv.typ, " != expected ", o.Type)
		assert.For(nv.val != nil, 41)
		s.data[o.Name] = nv.val
	}
}

func (s *stack) init() {
	s.vl = list.New()
}

func (s *stack) push(v *value) {
	assert.For(v != nil, 20)
	s.vl.PushFront(v)
}

func (s *stack) pop() (ret *value) {
	if s.vl.Len() > 0 {
		el := s.vl.Front()
		ret = s.vl.Remove(el).(*value)
	} else {
		halt.As(100, "pop on empty stack")
	}
	return
}

func (ctx *context) expr(_e ir.Expression, typ types.Type) {
	var eval func(ir.Expression)

	eval = func(_e ir.Expression) {
		//fmt.Println(_e, "for", typ)
		switch this := _e.(type) {
		case ir.EvaluatedExpression:
			eval(this.Eval())
		case *ir.NamedConstExpr:
			eval(this.Named.Expr)
		case *ir.AtomExpr:
			ctx.push(&value{typ: types.ATOM, val: Atom(this.Value)})
		case *ir.ConstExpr:
			switch typ {
			case types.INTEGER, types.BOOLEAN, types.CHAR, types.STRING, types.REAL, types.COMPLEX:
				ctx.push(cval(this))
			case types.TRILEAN:
				ctx.push(&value{typ: typ, val: tri.This(this.Value)})
			default:
				halt.As(100, "unknown target type ", typ)
			}
		case *ir.VariableExpr:
			ctx.findObj(this.Obj, func(v *value) *value {
				ctx.push(v)
				return nil
			})
		case *ir.SelectExpr:
			eval(this.Base)
			e := ctx.pop()
			ctx.sel(this.Sel, e, nil, func(v *value) *value {
				ctx.push(v)
				return nil
			})
		case *ir.Monadic:
			eval(this.Operand)
			v := ctx.pop()
			switch this.Op {
			case operation.Neg:
				switch v.typ {
				case types.INTEGER:
					i := v.toInt()
					i = i.Neg(i)
					ctx.push(&value{typ: v.typ, val: ThisInt(i)})
				case types.REAL:
					i := v.toReal()
					i = i.Neg(i)
					ctx.push(&value{typ: v.typ, val: ThisRat(i)})
				default:
					halt.As(100, "unknown type of operand ", v.typ)
				}
			case operation.Not:
				switch typ {
				case types.BOOLEAN:
					b := v.toBool()
					ctx.push(&value{typ: v.typ, val: !b})
				case types.TRILEAN:
					t := v.toTril()
					ctx.push(&value{typ: v.typ, val: tri.Not(t)})
				default:
					halt.As(100, "unexpected logical type")
				}
			case operation.Im:
				switch v.typ {
				case types.INTEGER:
					i := v.toInt()
					im := big.NewRat(0, 1)
					im.SetInt(i)
					re := big.NewRat(0, 1)
					c := &Cmp{}
					c.re = re
					c.im = im
					ctx.push(&value{typ: types.COMPLEX, val: c})
				case types.REAL:
					im := v.toReal()
					re := big.NewRat(0, 1)
					c := &Cmp{}
					c.re = re
					c.im = im
					ctx.push(&value{typ: types.COMPLEX, val: c})
				default:
					halt.As(100, "unexpected operand type ", v.typ)
				}
			default:
				halt.As(100, "unknown monadic op ", this.Op)
			}
		case *ir.Dyadic:
			var l, r *value
			if !(this.Op == operation.And || this.Op == operation.Or) {
				eval(this.Left)
				l = ctx.pop()
				eval(this.Right)
				r = ctx.pop()
				v := calcDyadic(l, this.Op, r)
				ctx.push(v)
			} else { //short boolean expr
				eval(this.Left)
				l = ctx.pop()
				switch this.Op {
				case operation.And:
					switch typ {
					case types.BOOLEAN:
						lb := l.toBool()
						if lb {
							eval(this.Right)
							r = ctx.pop()
							rb := r.toBool()
							lb = lb && rb
						}
						ctx.push(&value{typ: typ, val: lb})
					case types.TRILEAN:
						lt := l.toTril()
						if !tri.False(lt) {
							eval(this.Right)
							r = ctx.pop()
							rt := r.toTril()
							lt = tri.And(lt, rt)
						}
						ctx.push(&value{typ: typ, val: lt})
					default:
						halt.As(100, "unexpected logical type")
					}
				case operation.Or:
					switch typ {
					case types.BOOLEAN:
						lb := l.toBool()
						if !lb {
							eval(this.Right)
							r = ctx.pop()
							rb := r.toBool()
							lb = lb || rb
						}
						ctx.push(&value{typ: typ, val: lb})
					case types.TRILEAN:
						lt := l.toTril()
						if !tri.True(lt) {
							eval(this.Right)
							r = ctx.pop()
							rt := r.toTril()
							lt = tri.Or(lt, rt)
						}
						ctx.push(&value{typ: typ, val: lt})
					default:
						halt.As(100, "unexpected logical type")
					}
				default:
					halt.As(100, "unknown dyadic op ", this.Op)
				}
			}
		default:
			halt.As(100, "unknown expression ", reflect.TypeOf(this))
		}
	}
	eval(_e)
}

func (ctx *context) sel(_s ir.Selector, in, out *value, end func(*value) *value) {
	type hs func(*value, *value, ...hs) *value

	tail := func(l ...hs) (ret []hs) {
		if len(l) > 1 {
			ret = l[1:]
		}
		return
	}
	first := func(in, out *value, l ...hs) *value {
		if len(l) > 0 {
			fn := l[0]
			return fn(in, out, tail(l...)...)
		} else {
			return nil
		}
	}

	var ssel func(ir.Selector) []hs
	ssel = func(_s ir.Selector) (chain []hs) {
		switch s := _s.(type) {
		case ir.ChainSelector:
			for _, v := range s.Chain() {
				chain = append(chain, ssel(v)...)
			}
		case *ir.SelectVar:
			chain = append(chain, func(in, out *value, l ...hs) (ret *value) {
				//fmt.Println("select var ", s.Var, in, out)
				ctx.findObj(s.Var, func(val *value) *value {
					ret = first(in, val, l...)
					return ret
				})
				return
			})
		case *ir.SelectIndex:
			chain = append(chain, func(in, out *value, l ...hs) *value {
				//fmt.Println("select index ", in, out)
				ctx.expr(s.Expr, types.INTEGER)
				iv := ctx.pop()
				i := iv.toInt().Int64()
				if in != nil { //get
					switch in.typ {
					case types.STRING:
						buf := []rune(in.toStr())
						//fmt.Println(buf, i, buf[i])
						out = &value{typ: types.CHAR, val: buf[i]}
					default:
						halt.As(100, "unknown base type ", in.typ)
					}
					return first(in, out, l...)
				} else if out != nil { //set
					data := first(in, out, l...)
					//fmt.Println(data)
					switch out.typ {
					case types.STRING:
						buf := []rune(out.toStr())
						buf[i] = data.toRune()
						//fmt.Println(buf, i, buf[i])
						in = &value{typ: types.STRING, val: string(buf)}
					default:
						halt.As(100, "unknown base type ", out.typ)
					}
					return in
				} else {
					halt.As(100, "unexpected in/out state ", in, " ", out)
				}
				panic(0)
			})
		default:
			halt.As(100, " unknown selector ", reflect.TypeOf(s))
		}
		return
	}

	lh := ssel(_s)
	lh = append(lh, func(in, out *value, l ...hs) *value {
		return end(out)
	})
	first(in, out, lh...)
}

func (ctx *context) stmt(_s ir.Statement) {
	switch this := _s.(type) {
	case *ir.CallStmt:
		ctx.do(this.Proc)
	case *ir.AssignStmt:
		ctx.sel(this.Sel, nil, nil, func(in *value) *value {
			ctx.expr(this.Expr, in.typ)
			val := ctx.pop()
			return val
		})
	case *ir.IfStmt:
		done := false
		for _, i := range this.Cond {
			ctx.expr(i.Expr, types.BOOLEAN)
			val := ctx.pop()
			if val.toBool() {
				done = true
				for _, s := range i.Seq {
					ctx.do(s)
				}
				break
			}
		}
		if !done && this.Else != nil {
			for _, s := range this.Else.Seq {
				ctx.do(s)
			}
		}
	case *ir.WhileStmt:
		for stop := false; !stop; {
			stop = true
			for _, i := range this.Cond {
				ctx.expr(i.Expr, types.BOOLEAN)
				val := ctx.pop()
				if val.toBool() {
					stop = false
					for _, s := range i.Seq {
						ctx.do(s)
					}
					break
				}
			}
		}
	case *ir.RepeatStmt:
		for stop := false; !stop; {
			stop = false
			for _, s := range this.Cond.Seq {
				ctx.do(s)
			}
			ctx.expr(this.Cond.Expr, types.BOOLEAN)
			val := ctx.pop()
			stop = val.toBool()
		}
	default:
		halt.As(100, "unknown statement ", reflect.TypeOf(this))
	}
}

func (ctx *context) do(_t interface{}) {
	switch this := _t.(type) {
	case *ir.Module:
		ctx.alloc(this.VarDecl)
		if len(this.BeginSeq) > 0 {
			fmt.Println("BEGIN", this.Name, ctx.data)
			for _, v := range this.BeginSeq {
				ctx.do(v)
			}
		}
		if len(this.CloseSeq) > 0 {
			fmt.Println("CLOSE", ctx.data)
			for _, v := range this.CloseSeq {
				ctx.do(v)
			}
		}
		fmt.Println("END", this.Name, ctx.data)
	case *ir.Procedure:
		fmt.Println("BEGIN", this.Name)
		for _, v := range this.Seq {
			ctx.do(v)
		}
		fmt.Println("END", this.Name)
	case ir.Statement:
		ctx.stmt(this)
	default:
		halt.As(100, reflect.TypeOf(this))
	}
}

func (c *context) run() {
	c.do(c.root)
}

func connectTo(m *ir.Module) (ret *context) {
	ret = &context{}
	ret.root = m
	ret.storage.init()
	ret.stack.init()
	return
}

func run(m *ir.Module) {
	connectTo(m).run()
}

func init() {
	li.Run = run
}
