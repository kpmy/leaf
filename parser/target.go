package parser

import (
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
		fmt.Print("(")
		for _, v := range stack {
			if _, ok := v.e.(*exprBuilder); ok {
				fmt.Print(v.priority, " ")
				v.e.(ir.EvaluatedExpression).Eval()
			} else {
				fmt.Print(v.priority, reflect.TypeOf(v.e), " ")
			}
		}
		fmt.Print(")")
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
		switch root := r.e.(type) {
		case *ir.ConstExpr, *ir.NamedConstExpr, *ir.VariableExpr: //do nothing
			ret = stack
		case *ir.Monadic:
			expr, tail := first(stack)
			assert.For(expr != nil, 40)
			ok := false
			expr, ok = bypass(expr)
			if !ok {
				ret = trav(expr, tail)
			}
			root.Operand = expr.e
		case *ir.Dyadic:
			ret = stack
			ok := false

			right, ret := first(ret)
			assert.For(right != nil, 40)
			right, ok = bypass(right)
			if !ok {
				ret = trav(right, ret)
			}
			root.Right = right.e

			left, ret := first(ret)
			assert.For(left != nil, 40)
			left, ok = bypass(left)
			if !ok {
				ret = trav(left, ret)
			}
			root.Left = left.e
		case nil: //do nothing
		default:
			halt.As(100, "unsupported type ", reflect.TypeOf(root))
		}
		return
	}
	ok := false
	root, tail := first(stack)
	root, ok = bypass(root)
	if !ok {
		trav(root, tail)
	}
	ret = root.e
	return
}

func (e *exprBuilder) factor(expr ir.Expression) {
	e.stack = append(e.stack, &exprItem{e: expr, priority: highest})
}

func (e *exprBuilder) power(expr ir.Expression) {
	e.stack = append(e.stack, &exprItem{e: expr, priority: higher})
}

func (e *exprBuilder) product(expr ir.Expression) {
	e.stack = append(e.stack, &exprItem{e: expr, priority: high})
}

func (e *exprBuilder) quantum(expr ir.Expression) {
	e.stack = append(e.stack, &exprItem{e: expr, priority: normal})
}

func (e *exprBuilder) expr(expr ir.Expression) {
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

type blockBuilder struct {
	scope scopeLevel
	seq   []ir.Statement
}

func (b *blockBuilder) obj(id string) *ir.Variable {
	return b.scope.varScope[id]
}

func (b *blockBuilder) put(s ir.Statement) {
	b.seq = append(b.seq, s)
}
