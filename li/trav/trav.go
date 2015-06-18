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

type Int struct {
	big.Int
}

func NewInt(x int64) (ret *Int) {
	ret = &Int{}
	ret.Int = *big.NewInt(0)
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
		ret = &x.Int
	default:
		halt.As(100, "wrong int")
	}
	return
}

func (v *value) toBool() (ret bool) {
	switch x := v.val.(type) {
	case bool:
		ret = x
	default:
		halt.As(100, "wrong bool")
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
		default:
			halt.As(100, "unknown type ", v.Type)
		}
	}
}

func (s *storage) findObj(o *ir.Variable, fn func(*value) *value) {
	data := s.data[o.Name]
	nv := fn(&value{typ: o.Type, val: data})
	if nv != nil {
		assert.For(nv.typ == o.Type, 40, nv.typ, " != ", o.Type)
		assert.For(nv.val != nil, 41)
		s.data[o.Name] = nv.val
	}
}

func (s *stack) init() {
	s.vl = list.New()
}

func (s *stack) push(v *value) {
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
	const (
		lss = -1
		eq  = 0
		gtr = 1
	)

	var eval func(ir.Expression)

	eval = func(_e ir.Expression) {
		//fmt.Println(_e, "for", typ)
		switch this := _e.(type) {
		case ir.EvaluatedExpression:
			eval(this.Eval())
		case *ir.NamedConstExpr:
			eval(this.Named.Expr)
		case *ir.ConstExpr:
			switch typ {
			case types.INTEGER:
				c.push(&value{typ: this.Type, val: this.Value})
			case types.BOOLEAN:
				c.push(&value{typ: this.Type, val: this.Value})
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
				i := v.toInt()
				i = i.Neg(i)
				c.push(&value{typ: v.typ, val: ThisInt(i)})
			case operation.Not:
				switch typ {
				case types.BOOLEAN:
					b := v.toBool()
					c.push(&value{typ: v.typ, val: !b})
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
			} else { //short boolean expr
				eval(this.Left)
				l = c.pop()
			}
			switch this.Op {
			case operation.Sum:
				li := l.toInt()
				ri := r.toInt()
				x := li.Add(li, ri)
				c.push(&value{typ: l.typ, val: ThisInt(x)})
			case operation.Diff:
				li := l.toInt()
				ri := r.toInt()
				x := li.Sub(li, ri)
				c.push(&value{typ: l.typ, val: ThisInt(x)})
			case operation.Prod:
				li := l.toInt()
				ri := r.toInt()
				x := li.Mul(li, ri)
				c.push(&value{typ: l.typ, val: ThisInt(x)})
			case operation.Mod:
				li := l.toInt()
				ri := r.toInt()
				x := li.Mod(li, ri)
				c.push(&value{typ: l.typ, val: ThisInt(x)})
			case operation.Div:
				li := l.toInt()
				ri := r.toInt()
				x := li.Div(li, ri)
				c.push(&value{typ: l.typ, val: ThisInt(x)})
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
				default:
					halt.As(100, "unexpected logical type")
				}
			case operation.Lss:
				li := l.toInt()
				ri := r.toInt()
				res := li.Cmp(ri)
				c.push(&value{typ: typ, val: (res == lss)})
			case operation.Gtr:
				li := l.toInt()
				ri := r.toInt()
				res := li.Cmp(ri)
				c.push(&value{typ: typ, val: (res == gtr)})
			case operation.Leq:
				li := l.toInt()
				ri := r.toInt()
				res := li.Cmp(ri)
				c.push(&value{typ: typ, val: (res != gtr)})
			case operation.Geq:
				li := l.toInt()
				ri := r.toInt()
				res := li.Cmp(ri)
				c.push(&value{typ: typ, val: (res != lss)})
			case operation.Eq:
				li := l.toInt()
				ri := r.toInt()
				res := li.Cmp(ri)
				c.push(&value{typ: typ, val: (res == eq)})
			case operation.Neq:
				li := l.toInt()
				ri := r.toInt()
				res := li.Cmp(ri)
				c.push(&value{typ: typ, val: (res != eq)})
			default:
				halt.As(100, "unknown dyadic op ", this.Op)
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
