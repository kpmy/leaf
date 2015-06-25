package leap

import (
	"container/list"
	"fmt"
	"github.com/kpmy/ypk/assert"
	"github.com/kpmy/ypk/halt"
	"leaf/ir"
	"leaf/ir/modifiers"
	"reflect"
)

type block struct {
	parent    *block
	childList []*block
	cm        map[string]*ir.Const
	vm        map[string]*ir.Variable
	pm        map[string]*ir.Procedure
}

func (b *block) init() {
	b.cm = make(map[string]*ir.Const)
	b.vm = make(map[string]*ir.Variable)
	b.pm = make(map[string]*ir.Procedure)
}

func (b *block) this(id string) interface{} {
	c := b.cm[id]
	v := b.vm[id]
	p := b.pm[id]
	switch {
	case c != nil:
		return c
	case v != nil:
		return v
	case p != nil:
		return p
	}
	return nil
}

func (b *block) find(id string) (ret interface{}) {
	for x := b; x != nil && ret == nil; {
		ret = x.this(id)
		x = x.parent
	}
	return
}

type stack struct {
	ls *list.List
}

func (s *stack) init() {
	s.ls = list.New()
}

func (s *stack) push() {
	b := &block{}
	b.init()
	b.parent = s.this()
	s.ls.PushFront(b)
}

func (s *stack) pop() {
	if s.ls.Len() > 0 {
		b := s.this()
		if b.parent != nil {
			b.parent.childList = append(b.parent.childList, b)
		}
		s.ls.Remove(s.ls.Front())
	}
}

func (s *stack) this() (ret *block) {
	if s.ls.Len() > 0 {
		ret = s.ls.Front().Value.(*block)
	}
	return
}

type target struct {
	top *ir.Module
	st  *stack
}

func (t *target) init(mod string) {
	t.top = &ir.Module{Name: mod}
	t.top.Init()
	t.st = &stack{}
	t.st.init()
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
	name string
	sc   *block
}

func (e *forwardNamedConstExpr) Self() {}

func (e *forwardNamedConstExpr) Eval() (ret ir.Expression) {
	if c, _ := e.sc.find(e.name).(*ir.Const); c != nil {
		return &ir.NamedConstExpr{Named: c}
	} else {
		halt.As(100, "undefined constant ", e.name)
	}
	panic(0)
}

type forwardCall struct {
	name  string
	param []*forwardParam
	sc    *block
}

func (s *forwardCall) Do() {}
func (s *forwardCall) Fwd() ir.Statement {
	if p, _ := s.sc.find(s.name).(*ir.Procedure); p != nil {
		var param []*ir.Parameter
		for _, par := range s.param {
			x := &ir.Parameter{}
			x.Var = p.VarDecl[par.name]
			assert.For(x.Var != nil && x.Var.Modifier == modifiers.Semi, 30)
			x.Expr = par.expr
			param = append(param, x)
		}
		return &ir.CallStmt{Proc: p, Par: param}
	} else {
		halt.As(100, "undefined procedure ", s.name, s.sc)
	}
	panic(0)
}

type forwardParam struct {
	name string
	expr ir.Expression
	link ir.Selector
}

type exprBuilder struct {
	sc    *block
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
	if x := e.sc.find(id); x != nil {
		switch v := x.(type) {
		case *ir.Const:
			return &ir.NamedConstExpr{Named: v}
		case *ir.Variable:
			return &ir.VariableExpr{Obj: v}
		default:
			halt.As(100, "unexpected ", reflect.TypeOf(v))
		}
	} else {
		return &forwardNamedConstExpr{name: id, sc: e.sc}
	}
	panic(0)
}

func (b *exprBuilder) selector(sel ir.Selector) ir.Expression {
	return &ir.SelectExpr{Sel: sel}
}

type blockBuilder struct {
	sc  *block
	seq []ir.Statement
}

func (b *blockBuilder) isObj(id string) bool {
	x := b.sc.find(id)
	_, ok := x.(*ir.Procedure)
	return x != nil && !ok
}

func (b *blockBuilder) obj(id string) ir.Selector {
	_v := b.sc.find(id)
	assert.For(_v != nil, 30, id)
	v, ok := _v.(*ir.Variable)
	assert.For(ok, 31, id)
	return &ir.SelectVar{Var: v}
}

func (b *blockBuilder) call(id string, pl []*forwardParam) ir.Statement {
	if p, _ := b.sc.find(id).(*ir.Procedure); p != nil {
		var param []*ir.Parameter
		for _, par := range pl {
			x := &ir.Parameter{}
			x.Var = p.VarDecl[par.name]
			assert.For(x.Var != nil, 30)
			assert.For((par.expr != nil) != (par.link != nil), 31)
			if par.expr != nil {
				assert.For(x.Var.Modifier == modifiers.Semi, 32)
				x.Expr = par.expr
			} else {
				assert.For(x.Var.Modifier == modifiers.Full, 33)
				x.Sel = par.link
			}
			param = append(param, x)
		}
		return &ir.CallStmt{Proc: p, Par: param}
	} else {
		return &forwardCall{sc: b.sc, name: id, param: pl}
	}
}

func (b *blockBuilder) put(s ir.Statement) {
	b.seq = append(b.seq, s)
}

func (b *blockBuilder) decl(id string, p *ir.Procedure) {
	assert.For(b.sc.pm[id] == nil, 20)
	b.sc.pm[id] = p
}

type selBuilder struct {
	sc    *block
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

type constBuilder struct {
	sc *block
}

func (b *constBuilder) this(id string) (ret *ir.Const, free bool) {
	if ret = b.sc.cm[id]; ret != nil {
		free = false
		return
	} else if b.sc.parent != nil {
		x := b.sc.parent.this(id)
		return nil, x == nil
	}
	return nil, true
}

func (b *constBuilder) decl(id string, c *ir.Const) {
	assert.For(b.sc.cm[id] == nil, 20)
	b.sc.cm[id] = c
}

type varBuilder struct {
	sc *block
}

func (b *varBuilder) this(id string) (ret *ir.Variable, free bool) {
	if ret = b.sc.vm[id]; ret != nil {
		free = false
		return
	} else {
		x := b.sc.this(id)
		return nil, x == nil
	}
	return nil, true
}

func (b *varBuilder) decl(id string, v *ir.Variable) {
	assert.For(b.sc.vm[id] == nil, 20)
	b.sc.vm[id] = v
}
