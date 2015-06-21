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
	"leaf/li"
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
		v := rune(e.Value.(int32))
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
		default:
			halt.As(100, "unknown type ", v.Type)
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

func (c *context) expr(_e ir.Expression, typ types.Type) {
	var eval func(ir.Expression)

	eval = func(_e ir.Expression) {
		//fmt.Println(_e, "for", typ)
		switch this := _e.(type) {
		case ir.EvaluatedExpression:
			eval(this.Eval())
		case *ir.NamedConstExpr:
			eval(this.Named.Expr)
		case *ir.AtomExpr:
			c.push(&value{typ: types.ATOM, val: Atom(this.Value)})
		case *ir.ConstExpr:
			switch typ {
			case types.INTEGER, types.BOOLEAN, types.CHAR, types.STRING, types.REAL:
				c.push(cval(this))
			case types.TRILEAN:
				c.push(&value{typ: typ, val: tri.This(this.Value)})
			default:
				halt.As(100, "unknown target type ", typ)
			}
		case *ir.VariableExpr:
			c.findObj(this.Obj, func(v *value) *value {
				c.push(v)
				return nil
			})
		case *ir.Monadic:
			eval(this.Operand)
			v := c.pop()
			switch this.Op {
			case operation.Neg:
				switch v.typ {
				case types.INTEGER:
					i := v.toInt()
					i = i.Neg(i)
					c.push(&value{typ: v.typ, val: ThisInt(i)})
				case types.REAL:
					i := v.toReal()
					i = i.Neg(i)
					c.push(&value{typ: v.typ, val: ThisRat(i)})
				default:
					halt.As(100, "unknown type of operand ", v.typ)
				}
			case operation.Not:
				switch typ {
				case types.BOOLEAN:
					b := v.toBool()
					c.push(&value{typ: v.typ, val: !b})
				case types.TRILEAN:
					t := v.toTril()
					c.push(&value{typ: v.typ, val: tri.Not(t)})
				default:
					halt.As(100, "unexpected logical type")
				}
			default:
				halt.As(100, "unknown monadic op ", this.Op)
			}
		case *ir.Dyadic:
			var l, r *value
			if !(this.Op == operation.And || this.Op == operation.Or) {
				eval(this.Left)
				l = c.pop()
				eval(this.Right)
				r = c.pop()
				v := calcDyadic(l, this.Op, r)
				c.push(v)
			} else { //short boolean expr
				eval(this.Left)
				l = c.pop()
				switch this.Op {
				case operation.And:
					switch typ {
					case types.BOOLEAN:
						lb := l.toBool()
						if lb {
							eval(this.Right)
							r = c.pop()
							rb := r.toBool()
							lb = lb && rb
						}
						c.push(&value{typ: typ, val: lb})
					case types.TRILEAN:
						lt := l.toTril()
						if !tri.False(lt) {
							eval(this.Right)
							r = c.pop()
							rt := r.toTril()
							lt = tri.And(lt, rt)
						}
						c.push(&value{typ: typ, val: lt})
					default:
						halt.As(100, "unexpected logical type")
					}
				case operation.Or:
					switch typ {
					case types.BOOLEAN:
						lb := l.toBool()
						if !lb {
							eval(this.Right)
							r = c.pop()
							rb := r.toBool()
							lb = lb || rb
						}
						c.push(&value{typ: typ, val: lb})
					case types.TRILEAN:
						lt := l.toTril()
						if !tri.True(lt) {
							eval(this.Right)
							r = c.pop()
							rt := r.toTril()
							lt = tri.Or(lt, rt)
						}
						c.push(&value{typ: typ, val: lt})
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

func (c *context) stmt(_s ir.Statement) {
	switch this := _s.(type) {
	case *ir.AssignStmt:
		c.expr(this.Expr, this.Object.Type)
		val := c.pop()
		c.findObj(this.Object, func(*value) *value {
			return val
		})
	default:
		halt.As(100, "unknown statement ", reflect.TypeOf(this))
	}
}

func (c *context) do(_t interface{}) {
	switch this := _t.(type) {
	case *ir.Module:
		c.alloc(this.VarDecl)
		fmt.Println("BEGIN", this.Name, c.data)
		for _, v := range this.BeginSeq {
			c.do(v)
		}
		for _, v := range this.CloseSeq {
			c.do(v)
		}
		fmt.Println("CLOSE", c.data)
	case ir.Statement:
		c.stmt(this)
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
