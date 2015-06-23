package parser

import (
	"fmt"
	"github.com/kpmy/ypk/assert"
	"github.com/kpmy/ypk/halt"
	"leaf/ir"
	"reflect"
)

type target struct {
	root *ir.Module
}

func (t *target) init(mod string) {
	t.root = &ir.Module{Name: mod}
	t.root.Init()
}

type scopeLevel struct {
	varScope   map[string]*ir.Variable
	constScope map[string]*ir.Const
	procScope  map[string]*ir.Procedure
}

type level int

const (
	low level = iota
	normal
	high
	higher
	highest
)

type exprItem struct {
	e        ir.Expression
	priority level
}

type forwardNamedConstExpr struct {
	name  string
	scope map[string]*ir.Const
}

func (e *forwardNamedConstExpr) Self() {}

func (e *forwardNamedConstExpr) Eval() (ret ir.Expression) {
	if c := e.scope[e.name]; c != nil {
		return &ir.NamedConstExpr{Named: c}
	} else {
		halt.As(100, "undefined constant ", e.name)
	}
	panic(0)
}

type forwardCall struct {
	name  string
	scope map[string]*ir.Procedure
}

func (s *forwardCall) Do() {}
func (s *forwardCall) Fwd() ir.Statement {
	if p := s.scope[s.name]; p != nil {
		return &ir.CallStmt{Proc: p}
	} else {
		halt.As(100, "undefined procedure ", s.name)
	}
	panic(0)
}

type exprBuilder struct {
	scope scopeLevel
	stack []*exprItem
}

func (e *exprBuilder) Self() {}

func (e *exprBuilder) Eval() (ret ir.Expression) {
	assert.For(len(e.stack) > 0, 20)
	last := func(s []*exprItem) []*exprItem {
		if len(s) > 1 {
			return s[1:]
		}
		return nil
	}
	first := func(s []*exprItem) (ret *exprItem, tail []*exprItem) {
		if len(s) > 0 {
			ret = s[0]
			tail = s
		}
		tail = last(tail)
		return
	}
	//invert stack
	var stack []*exprItem
	for _, v := range e.stack {
		tmp := stack
		stack = nil
		stack = append(stack, v)
		stack = append(stack, tmp...)
	}
	/*
		{
			fmt.Print("(")
			for _, v := range stack {
				if _, ok := v.e.(*exprBuilder); ok {
					fmt.Print(v.priority, " ")
					v.e.(ir.EvaluatedExpression).Eval()
				} else {
					fmt.Print(v.priority, reflect.TypeOf(v.e), " ")
				}
			}
			fmt.Println(")")
		}
	*/
	var trav func(*exprItem, []*exprItem) []*exprItem
	bypass := func(expr *exprItem) (ret *exprItem, skip bool) {
		ret = expr
		if b, ok := ret.e.(*exprBuilder); ok {
			skip = true
			ret = &exprItem{e: b.Eval(), priority: ret.priority}
		} else if f, ok := ret.e.(*forwardNamedConstExpr); ok {
			skip = true
			ret = &exprItem{e: f.Eval(), priority: ret.priority}
		}
		return
	}
	trav = func(r *exprItem, stack []*exprItem) (ret []*exprItem) {
		//fmt.Println(reflect.TypeOf(r.e))
		switch root := r.e.(type) {
		case *ir.ConstExpr, *ir.NamedConstExpr, *ir.VariableExpr, *ir.SelectExpr: //do nothing
			ret = stack
		case *ir.Monadic:
			expr, tail := first(stack)
			assert.For(expr != nil, 40)
			ok := false
			expr, ok = bypass(expr)
			if !ok {
				//fmt.Println("mop trav")
				ret = trav(expr, tail)
			}
			root.Operand = expr.e
		case *ir.Dyadic:
			ret = stack
			ok := false

			right, tail := first(ret)
			assert.For(right != nil, 40)
			right, ok = bypass(right)
			if !ok {
				//fmt.Println("dop right trav")
				ret = trav(right, tail)
			}
			root.Right = right.e

			left, tail := first(ret)
			assert.For(left != nil, 40)
			left, ok = bypass(left)
			if !ok {
				//fmt.Println("dop left trav")
				ret = trav(left, tail)
			}
			root.Left = left.e
		case nil: //do nothing
		default:
			halt.As(100, "unsupported type ", fmt.Sprint(reflect.TypeOf(root)))
		}
		return
	}
	ok := false
	root, tail := first(stack)
	root, ok = bypass(root)
	if !ok {
		//fmt.Println("root trav")
		trav(root, tail)
	}
	ret = root.e
	{
		var eprint func(ir.Expression)
		eprint = func(_e ir.Expression) {
			switch e := _e.(type) {
			case *ir.ConstExpr:
				fmt.Println(e.Value)
			case *ir.NamedConstExpr:
				fmt.Println(e.Named)
			case *ir.VariableExpr:
				fmt.Print("$", e.Obj.Name)
				fmt.Println()
			case *ir.Monadic:
				fmt.Println("mop")
				eprint(e.Operand)
				fmt.Println(e.Op)
			case *ir.Dyadic:
				fmt.Println("dop left")
				eprint(e.Left)
				fmt.Println("dop right")
				eprint(e.Right)
				fmt.Println(e.Op)
			default:
				halt.As(100, reflect.TypeOf(e))
			}
		}
		//fmt.Println("evaluated")
		//eprint(ret)
	}
	return
}

func (e *exprBuilder) factor(expr ir.Expression) {
	//fmt.Println("factor ", reflect.TypeOf(expr), expr)
	e.stack = append(e.stack, &exprItem{e: expr, priority: highest})
}

func (e *exprBuilder) power(expr ir.Expression) {
	//fmt.Println("power ", reflect.TypeOf(expr), expr)
	e.stack = append(e.stack, &exprItem{e: expr, priority: higher})
}

func (e *exprBuilder) product(expr ir.Expression) {
	//fmt.Println("product ", reflect.TypeOf(expr), expr)
	e.stack = append(e.stack, &exprItem{e: expr, priority: high})
}

func (e *exprBuilder) quantum(expr ir.Expression) {
	//fmt.Println("quantum ", reflect.TypeOf(expr), expr)
	e.stack = append(e.stack, &exprItem{e: expr, priority: normal})
}

func (e *exprBuilder) expr(expr ir.Expression) {
	//fmt.Println("expr ", reflect.TypeOf(expr), expr)
	e.stack = append(e.stack, &exprItem{e: expr, priority: low})
}

func (e *exprBuilder) as(id string) ir.Expression {
	if e.scope.constScope != nil && e.scope.varScope != nil {
		if c := e.scope.constScope[id]; c != nil {
			return &ir.NamedConstExpr{Named: c}
		} else if v := e.scope.varScope[id]; v != nil {
			return &ir.VariableExpr{Obj: v}
		}
	} else if e.scope.constScope != nil && e.scope.varScope == nil {
		if c := e.scope.constScope[id]; c != nil {
			return &ir.NamedConstExpr{Named: c}
		} else {
			return &forwardNamedConstExpr{name: id, scope: e.scope.constScope}
		}
	}
	panic(0)
}

func (b *exprBuilder) selector(sel ir.Selector) ir.Expression {
	return &ir.SelectExpr{Sel: sel}
}

type blockBuilder struct {
	scope    scopeLevel
	seq      []ir.Statement
	procList []*ir.Procedure
}

func (b *blockBuilder) isObj(id string) bool {
	return b.scope.varScope[id] != nil
}

func (b *blockBuilder) obj(id string) ir.Selector {
	v := b.scope.varScope[id]
	assert.For(v != nil, 30)
	return &ir.SelectVar{Var: v}
}

func (b *blockBuilder) call(id string) ir.Statement {
	if p := b.scope.procScope[id]; p != nil {
		return &ir.CallStmt{Proc: p}
	} else {
		return &forwardCall{scope: b.scope.procScope, name: id}
	}
}

func (b *blockBuilder) put(s ir.Statement) {
	b.seq = append(b.seq, s)
}

func (b *blockBuilder) putProc(p *ir.Procedure) {
	b.procList = append(b.procList, p)
}

type selBuilder struct {
	scope scopeLevel
	chain []ir.Selector
}

func (s *selBuilder) Select() {}

func (s *selBuilder) head(obj ir.Selector) {
	tmp := s.chain
	s.chain = nil
	s.chain = append(s.chain, obj)
	s.chain = append(s.chain, tmp...)
}

func (s *selBuilder) join(obj ir.Selector) {
	s.chain = append(s.chain, obj)
}

func (s *selBuilder) appy(expr ir.Expression) ir.Expression {
	if len(s.chain) > 0 {
		ret := &ir.SelectExpr{}
		ret.Base = expr
		ret.Sel = s
		return ret
	} else {
		return expr
	}
}

func (s *selBuilder) Chain() []ir.Selector {
	assert.For(len(s.chain) > 0, 20)
	return s.chain
}
