package leap

import (
	"container/list"
	"fmt"
	"github.com/kpmy/ypk/assert"
	"github.com/kpmy/ypk/fn"
	"github.com/kpmy/ypk/halt"
	"leaf/ir"
	"leaf/ir/modifiers"
	"leaf/lenin/rt"
	"reflect"
)

type block struct {
	parent    *block
	childList []*block
	cm        map[string]*ir.Const
	vm        map[string]*ir.Variable
	pm        map[string]*ir.Procedure
	in        []*ir.Variable
	pre, post []ir.Expression
	il        []*ir.Import
	im        map[string]*ir.Import
}

func (b *block) init() {
	b.cm = make(map[string]*ir.Const)
	b.vm = make(map[string]*ir.Variable)
	b.pm = make(map[string]*ir.Procedure)

	b.im = make(map[string]*ir.Import)
}

func (b *block) imp(id string) (imp *ir.Import) {
	imp = b.im[id]
	if imp == nil && b.parent != nil {
		imp = b.parent.imp(id)
	}
	return
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

type imported struct {
	top *ir.Import
}

func (i *imported) init(mod string) {
	i.top = &ir.Import{Name: mod}
	i.top.Init()
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

type forwardInfix struct {
	name string
	sc   *block
	args int
}

func (e *forwardInfix) Self() {}

func (e *forwardInfix) Eval() (ret ir.Expression) {
	if p, _ := e.sc.find(e.name).(*ir.Procedure); p != nil {
		assert.For(len(p.Infix)-1 == e.args, 20)
		i := &ir.Infix{Proc: p, Len: e.args}
		return i
	} else if p := rt.StdImp.ProcDecl[e.name]; p != nil {
		assert.For(len(p.This().Infix)-1 == e.args, 20)
		i := &ir.InvokeInfix{Mod: rt.StdImp.Name, Proc: p.Name(), Len: e.args}
		return i
	} else {
		halt.As(100, "undefined procedure ", rt.StdImp.Name, ".", e.name)
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
	params := func(p *ir.Procedure) (param []*ir.Parameter) {
		for _, par := range s.param {
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
		return
	}
	if p, _ := s.sc.find(s.name).(*ir.Procedure); p != nil {
		return &ir.CallStmt{Proc: p, Par: params(p)}
	} else if p := rt.StdImp.ProcDecl[s.name]; p != nil {
		return &ir.InvokeStmt{Mod: rt.StdImp.Name, Proc: p.Name(), Par: params(p.This())}
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
		} else if i, ok := ret.e.(*forwardInfix); ok {
			ret = &exprItem{e: i.Eval(), priority: ret.priority}
			_, skip = ret.e.(*ir.Infix)
		}
		return
	}
	trav = func(r *exprItem, stack []*exprItem) (ret []*exprItem) {
		//fmt.Println(reflect.TypeOf(r.e))
		switch root := r.e.(type) {
		case *ir.ConstExpr, *ir.NamedConstExpr, *ir.VariableExpr, *ir.SelectExpr, *ir.SetExpr, *ir.ListExpr, *ir.MapExpr: //do nothing
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
		case *ir.TypeTest:
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
		case *ir.Infix:
			ret = stack
			ok := false
			//fmt.Println(root.Len)
			for i := 0; i < root.Len; i++ {
				expr, tail := first(ret)
				assert.For(expr != nil, 40)
				//fmt.Println("NOW", expr, len(ret))
				expr, ok = bypass(expr)
				if !ok {
					//fmt.Println("trav")
					ret = trav(expr, tail)
				} else {
					ret = tail
				}
				root.Args = append(root.Args, expr.e)
			}
			//reverse
			tmp := root.Args
			root.Args = nil
			for i := len(tmp) - 1; i >= 0; i-- {
				root.Args = append(root.Args, tmp[i])
			}
		case *ir.InvokeInfix:
			ret = stack
			ok := false
			//fmt.Println(root.Len)
			for i := 0; i < root.Len; i++ {
				expr, tail := first(ret)
				assert.For(expr != nil, 40)
				//fmt.Println("NOW", expr, len(ret))
				expr, ok = bypass(expr)
				if !ok {
					//fmt.Println("trav")
					ret = trav(expr, tail)
				} else {
					ret = tail
				}
				root.Args = append(root.Args, expr.e)
			}
			//reverse
			tmp := root.Args
			root.Args = nil
			for i := len(tmp) - 1; i >= 0; i-- {
				root.Args = append(root.Args, tmp[i])
			}
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
			case *ir.Infix:
				fmt.Println("infix")
				for _, x := range e.Args {
					eprint(x)
				}
			case *ir.InvokeInfix:
				fmt.Println("invoke")
				for _, x := range e.Args {
					eprint(x)
				}
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
	//	fmt.Println("factor ", reflect.TypeOf(expr), expr)
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

func (e *exprBuilder) asImp(mid, id string) (ret ir.Expression) {
	imp := e.sc.imp(mid)
	assert.For(imp != nil, 40)
	if x := imp.ConstDecl[id]; x != nil {
		c := x.This()
		ret = &ir.NamedConstExpr{Named: c}
	} else if x := imp.VarDecl[id]; x != nil {
		v := x.This()
		ret = &ir.VariableExpr{Obj: v}
	} else {
		halt.As(100, "unknown import ", mid, ".", id)
	}
	return
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

func (b *exprBuilder) infix(mid, id string, num int) ir.Expression {
	var p *ir.Procedure
	mod := ""
	if mid != "" {
		imp := b.sc.im[mid]
		if x := imp.ProcDecl[id]; x != nil {
			p = x.This()
			mod = imp.Name
		} else {
			halt.As(100, "unknown import ", mid, ".", id)
		}
	} else {
		if p, _ = b.sc.find(id).(*ir.Procedure); p == nil {
			return &forwardInfix{name: id, sc: b.sc, args: num}
		}
	}
	assert.For(p != nil, 40)
	assert.For(len(p.Infix)-1 == num, 20, len(p.Infix), num)
	i := &ir.Infix{Mod: mod, Proc: p, Len: num}
	return i
}

type blockBuilder struct {
	sc  *block
	seq []ir.Statement
}

func (b *blockBuilder) impObj(mid, id string) ir.Selector {
	imp := b.sc.imp(mid)
	assert.For(imp != nil, 40)
	var v *ir.Variable
	if x := imp.VarDecl[id]; x != nil {
		v = x.This()
		return &ir.SelectVar{Var: v}

	}
	return nil
}

func (b *blockBuilder) obj(id string) (ret ir.Selector) {
	_v := b.sc.find(id)
	if _v != nil {
		v, ok := _v.(*ir.Variable)
		if ok {
			ret = &ir.SelectVar{Var: v}
		}
	}
	return
}

func (b *blockBuilder) call(mid, id string, pl []*forwardParam) ir.Statement {
	var p *ir.Procedure
	mod := ""
	if mid != "" {
		imp := b.sc.im[mid]
		if x := imp.ProcDecl[id]; x != nil {
			p = x.This()
			mod = imp.Name
		} else {
			halt.As(100, "unknown import ", mid, ".", id)
		}
	} else {
		if p, _ = b.sc.find(id).(*ir.Procedure); p == nil {
			return &forwardCall{sc: b.sc, name: id, param: pl}
		}
	}
	assert.For(p != nil, 40)
	var param []*ir.Parameter
	for _, par := range pl {
		x := &ir.Parameter{}
		if len(p.VarDecl) > 0 {
			x.Var = p.VarDecl[par.name]
		} else { //shortcut for recursion
			x.Var = b.sc.vm[par.name]
		}
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
	return &ir.CallStmt{Mod: mod, Proc: p, Par: param}

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

func (s *selBuilder) apply(before ir.Selector, expr ir.Expression) ir.Expression {
	if x, ok := before.(*selBuilder); ok {
		if len(x.chain) == 0 {
			before = nil
		}
	}
	if len(s.chain) > 0 || !fn.IsNil(before) {
		ret := &ir.SelectExpr{Before: before}
		ret.Base = expr
		if len(s.chain) > 0 {
			ret.After = s
		}
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
