package trav

import (
	"container/list"
	"fmt"
	"github.com/kpmy/trigo"
	"github.com/kpmy/ypk/assert"
	"github.com/kpmy/ypk/fn"
	"github.com/kpmy/ypk/halt"
	"leaf/ir"
	"leaf/ir/modifiers"
	"leaf/ir/operation"
	"leaf/ir/types"
	"leaf/lenin"
	"leaf/lenin/rt"
	"math/big"
	"reflect"
)

type context struct {
	data *storeStack
	exprStack
	load []*ir.Module
	tgt  *storage
}

type anyData interface {
	String() string
	read() interface{}
	write(interface{})
}

type direct struct {
	____x interface{}
}

func (d *direct) String() string {
	return fmt.Sprint(d.____x)
}

func (d *direct) set(x interface{}) {
	_, fake := x.(*value)
	assert.For(!fake, 21)
	d.____x = x
}

func (d *direct) read() interface{} { return d.____x }
func (d *direct) write(x interface{}) {
	assert.For(x != nil, 20)
	d.set(x)
}

type indirect struct {
	sel   ir.Selector
	ctx   *context
	stor  *storage
	____x interface{}
}

func (i *indirect) String() string {
	if i.sel != nil {
		return fmt.Sprint("@", i.sel)
	} else {
		return fmt.Sprint("@", i.____x)
	}
}
func (d *indirect) set(x interface{}) {
	_, fake := x.(*value)
	assert.For(!fake, 21)
	d.____x = x
}

func (d *indirect) doSel(in, out *value, end func(*value) *value) {
	d.ctx.sel(d.sel, in, out, end)
}

func (d *indirect) read() (ret interface{}) {
	if d.sel != nil {
		//fmt.Println("indirect")
		d.stor.lock = &lock{}
		d.doSel(nil, nil, func(v *value) *value {
			ret = v.val
			return nil
		})
		d.stor.lock = nil
		_, fake := ret.(*value)
		assert.For(!fake, 21)
		return
	} else {
		return d.____x
	}
}

func (d *indirect) write(x interface{}) {
	assert.For(x != nil, 20)
	_, fake := x.(*value)
	assert.For(!fake, 21)
	if d.sel != nil {
		d.stor.lock = &lock{}
		d.doSel(nil, nil, func(v *value) *value {
			return &value{typ: v.typ, val: x}
		})
		d.stor.lock = nil
	} else {
		d.set(x)
	}
}

type storage struct {
	root   *ir.Module
	link   interface{}
	schema map[string]*ir.Variable
	data   map[string]anyData
	lock   *lock
	prev   *storage
}

type lock struct{}

type storeStack struct {
	store map[string]*storage
	sl    *list.List
	ml    *list.List
	owner *context
}

func (s *storeStack) init(o *context) {
	s.store = make(map[string]*storage)
	s.sl = list.New()
	s.ml = list.New()
	s.owner = o
}

func (s *storeStack) mpush(m *ir.Module) {
	assert.For(m != nil, 20)
	s.ml.PushFront(m)
}

func (s *storeStack) mtop() *ir.Module {
	if s.ml.Len() > 0 {
		return s.ml.Front().Value.(*ir.Module)
	}
	return nil
}

func (s *storeStack) mpop() (ret *ir.Module) {
	if s.ml.Len() > 0 {
		el := s.ml.Front()
		ret = s.ml.Remove(el).(*ir.Module)
	} else {
		halt.As(100, "pop on empty stack")
	}
	return
}

func (s *storeStack) push(st *storage) {
	assert.For(st != nil, 20)
	st.prev = s.top()
	s.sl.PushFront(st)
}

func (s *storeStack) pop() (ret *storage) {
	if s.sl.Len() > 0 {
		el := s.sl.Front()
		ret = s.sl.Remove(el).(*storage)
	} else {
		halt.As(100, "pop on empty stack")
	}
	return
}

func (s *storeStack) top() *storage {
	if s.sl.Len() > 0 {
		return s.sl.Front().Value.(*storage)
	}
	return nil
}

func (s *storeStack) alloc(_x interface{}) {
	switch x := _x.(type) {
	case *ir.Module:
		d := &storage{root: x}
		d.alloc(x.VarDecl)
		s.store[x.Name] = d
		s.mpush(x)
		fmt.Println("alloc", x.Name, d.data)
	case *ir.Procedure:
		d := &storage{root: s.mtop(), link: x}
		d.alloc(x.VarDecl)
		s.push(d)
		fmt.Println("alloc", x.Name, d.data)
	case ir.ImportProcedure:
		d := &storage{root: s.mtop(), link: x}
		d.alloc(x.This().VarDecl)
		s.push(d)
		fmt.Println("alloc", x.Name(), d.data)
	default:
		halt.As(100, reflect.TypeOf(x))
	}
}

func (s *storeStack) dealloc(_x interface{}) (ret *storage) {
	switch x := _x.(type) {
	case *ir.Module:
		fmt.Println("dealloc", x.Name, s.store[x.Name].data)
		s.store[x.Name] = nil
		//TODO проверить наличие связанных элементов стека
	case *ir.Procedure:
		assert.For(s.top() != nil && s.top().link == x, 20)
		fmt.Println("dealloc", x.Name, s.top().data)
		ret = s.top()
		s.pop()
	case ir.ImportProcedure:
		assert.For(s.top() != nil && s.top().link == x, 20)
		fmt.Println("dealloc", x.Name(), s.top().data)
		ret = s.top()
		s.pop()
	default:
		halt.As(100, reflect.TypeOf(x))
	}
	return
}

func (s *storage) init() {
	s.schema = make(map[string]*ir.Variable)
	s.data = make(map[string]anyData)
}

func (s *storage) alloc(vl map[string]*ir.Variable) {
	assert.For(vl != nil, 20)
	s.schema = vl
	s.data = make(map[string]anyData)
	for _, v := range s.schema {

		init := func(val interface{}) (ret anyData) {
			assert.For(val != nil, 20)
			switch v.Modifier {
			case modifiers.Full:
				x := &indirect{stor: s}
				x.set(val)
				ret = x
			case modifiers.Semi, modifiers.None:
				x := &direct{}
				x.set(val)
				ret = x
			default:
				halt.As(100, "wrong modifier ", v.Modifier)
			}
			return
			panic(0)
		}

		switch v.Type {
		case types.INTEGER:
			s.data[v.Name] = init(NewInt(0))
		case types.BOOLEAN:
			s.data[v.Name] = init(false)
		case types.TRILEAN:
			s.data[v.Name] = init(tri.NIL)
		case types.CHAR:
			s.data[v.Name] = init(rune(0))
		case types.STRING:
			s.data[v.Name] = init("")
		case types.ATOM:
			s.data[v.Name] = init(Atom(""))
		case types.REAL:
			s.data[v.Name] = init(NewRat(0.0))
		case types.COMPLEX:
			s.data[v.Name] = init(NewCmp(0.0, 0.0))
		case types.ANY:
			s.data[v.Name] = init(&Any{})
		default:
			halt.As(100, "unknown type ", v.Name, ": ", v.Type)
		}
	}
}

func (s *storage) List() (ret []*ir.Variable) {
	for _, x := range s.schema {
		ret = append(ret, x)
	}
	return
}

func (s *storage) Get(name string) interface{} {
	if d := s.data[name]; d != nil {
		return d.read()
	} else {
		halt.As(100, "object not found")
	}
	panic(0)
}

func (s *storage) Set(name string, x interface{}) {
	assert.For(x != nil, 20)
	if d := s.data[name]; d != nil {
		d.write(x)
	} else {
		halt.As(100, "object not found")
	}
}

func (st *storeStack) ref(o *ir.Variable, sel ir.Selector) {
	find := func(s *storage) (ret bool) {
		if data, ok := s.data[o.Name]; ok {
			assert.For(data != nil, 20)
			r := data.(*indirect)
			r.sel = sel
			r.ctx = st.owner
			ret = true
		}
		return
	}
	found := false
	if local := st.top(); local != nil {
		found = find(local)
	}
	assert.For(found, 60)
}

func (s *storeStack) inner(o *ir.Variable, fn func(*value) *value) {
	find := func(s *storage) (ret bool) {
		if data, ok := s.data[o.Name]; ok {
			assert.For(data != nil, 20)
			nv := fn(&value{typ: o.Type, val: data.read()})
			if nv != nil {
				assert.For(compTypes(nv.typ, o.Type), 40, "provided ", nv.typ, " != expected ", o.Type)
				nv = conv(nv, o.Type)
				s.data[o.Name].write(nv.val)
				fmt.Println("touch", o.Name, nv.val)
			}
			ret = true
		}
		return
	}
	found := false
	for local := s.top(); local != nil; {
		if local.lock != nil {
			//fmt.Println("locked, try prev")
			local = local.prev
		} else {
			found = find(local)
			break
		}
	}
	if !found {
		mod := s.mtop()
		found = find(s.store[mod.Name])
	}
	assert.For(found, 60, `"`, o.Name, `"`)
}

func (s *storeStack) outer(st *storage, o *ir.Variable, fn func(*value) *value) {
	find := func(s *storage) (ret bool) {
		if data, ok := s.data[o.Name]; ok {
			assert.For(data != nil, 20)
			nv := fn(&value{typ: o.Type, val: data.read()})
			if nv != nil {
				assert.For(compTypes(nv.typ, o.Type), 40, "provided ", nv.typ, " != expected ", o.Type)
				nv = conv(nv, o.Type)
				s.data[o.Name].write(nv.val)
				fmt.Println("touch", o.Name, nv.val)
			}
			ret = true
		}
		return
	}
	if st != nil {
		found := find(st)
		assert.For(found, 60)
	} else {
		s.inner(o, fn)
		return
	}
}

type exprStack struct {
	vl *list.List
}

func (s *exprStack) init() {
	s.vl = list.New()
}

func (s *exprStack) push(v *value) {
	assert.For(v != nil, 20)
	_, fake := v.val.(*value)
	assert.For(!fake, 21)
	s.vl.PushFront(v)
}

func (s *exprStack) pop() (ret *value) {
	if s.vl.Len() > 0 {
		el := s.vl.Front()
		ret = s.vl.Remove(el).(*value)
	} else {
		halt.As(100, "pop on empty stack")
	}
	return
}

func (ctx *context) expr(_e ir.Expression) {
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
			ctx.push(cval(this))
		case *ir.VariableExpr:
			scope := ctx.tgt
			ctx.tgt = nil
			ctx.data.outer(scope, this.Obj, func(v *value) *value {
				ctx.push(v)
				return nil
			})
		case *ir.SelectExpr:
			if !fn.IsNil(this.Before) {
				ctx.sel(this.Before, nil, nil, func(v *value) *value { return nil })
			}
			eval(this.Base)
			if !fn.IsNil(this.After) {
				e := ctx.pop()
				ctx.sel(this.After, e, nil, func(v *value) *value {
					ctx.push(v)
					return nil
				})
			}
		case *ir.TypeTest:
			eval(this.Operand)
			v := ctx.pop()
			a := v.toAny()
			switch {
			case a.x == nil:
				ctx.push(&value{typ: types.TRILEAN, val: tri.NIL})
			case a.x != nil && a.typ == this.Typ:
				ctx.push(&value{typ: types.TRILEAN, val: tri.TRUE})
			case a.x != nil && a.typ != this.Typ:
				ctx.push(&value{typ: types.TRILEAN, val: tri.FALSE})
			default:
				halt.As(100, "unhandled type testing for ", v.typ, v.val)
			}
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
				switch v.typ {
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
					switch l.typ {
					case types.BOOLEAN:
						lb := l.toBool()
						if lb {
							eval(this.Right)
							r = ctx.pop()
							rb := r.toBool()
							lb = lb && rb
						}
						ctx.push(&value{typ: l.typ, val: lb})
					case types.TRILEAN:
						lt := l.toTril()
						if !tri.False(lt) {
							eval(this.Right)
							r = ctx.pop()
							rt := r.toTril()
							lt = tri.And(lt, rt)
						}
						ctx.push(&value{typ: l.typ, val: lt})
					default:
						halt.As(100, "unexpected logical type")
					}
				case operation.Or:
					switch l.typ {
					case types.BOOLEAN:
						lb := l.toBool()
						if !lb {
							eval(this.Right)
							r = ctx.pop()
							rb := r.toBool()
							lb = lb || rb
						}
						ctx.push(&value{typ: l.typ, val: lb})
					case types.TRILEAN:
						lt := l.toTril()
						if !tri.True(lt) {
							eval(this.Right)
							r = ctx.pop()
							rt := r.toTril()
							lt = tri.Or(lt, rt)
						}
						ctx.push(&value{typ: l.typ, val: lt})
					default:
						halt.As(100, "unexpected logical type")
					}
				default:
					halt.As(100, "unknown dyadic op ", this.Op)
				}
			}
		case *ir.Infix:
			var vl []*value
			for _, e := range this.Args {
				eval(e)
				val := ctx.pop()
				vl = append(vl, val)
			}
			assert.For(len(vl) == len(this.Proc.Infix)-1, 40, len(vl), len(this.Proc.Infix))
			var pl []interface{}
			for i, v := range vl {
				par := this.Proc.Infix[i+1]
				p := &param{obj: par, val: v}
				//fmt.Println(par.Name, vl[i].val)
				pl = append(pl, p)
			}
			if this.Mod != "" {
				top := ctx.data.store[this.Mod]
				ctx.data.mpush(top.root)
			}
			if x, _ := ctx.do(this.Proc, pl...).(*storage); x != nil {
				if this.Mod != "" {
					top := ctx.data.mpop()
					assert.For(top.Name == this.Mod, 60)
				}
				out := this.Proc.Infix[0]
				val := x.data[out.Name]
				assert.For(val != nil, 40)
				ctx.push(&value{typ: out.Type, val: val.read()})
			} else {
				halt.As(100, "no result from infix")
			}
		case *ir.InvokeInfix:
			assert.For(rt.StdImp.Name == this.Mod, 20)
			proc := rt.StdImp.ProcDecl[this.Proc].This()
			var vl []*value
			for _, e := range this.Args {
				eval(e)
				val := ctx.pop()
				vl = append(vl, val)
			}
			assert.For(len(vl) == len(proc.Infix)-1, 40, len(vl), len(proc.Infix))
			var pl []interface{}
			for i, v := range vl {
				par := proc.Infix[i+1]
				p := &param{obj: par, val: v}
				//fmt.Println(par.Name, vl[i].val)
				pl = append(pl, p)
			}
			if x, _ := ctx.invoke(this.Mod, this.Proc, pl...).(*storage); x != nil {
				out := proc.Infix[0]
				val := x.data[out.Name]
				assert.For(val != nil, 40)
				ctx.push(&value{typ: out.Type, val: val.read()})
			} else {
				halt.As(100, "no result from infix")
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
		case *ir.SelectMod:
			chain = append(chain, func(in, out *value, l ...hs) (ret *value) {
				ctx.tgt = ctx.data.store[s.Mod]
				return first(in, out, l...)
			})
		case *ir.SelectVar:
			chain = append(chain, func(in, out *value, l ...hs) (ret *value) {
				//fmt.Println("select var ", s.Var, in, out)
				scope := ctx.tgt
				ctx.tgt = nil
				ctx.data.outer(scope, s.Var, func(val *value) *value {
					ret = first(in, val, l...)
					return ret
				})
				return
			})
		case *ir.SelectIndex:
			chain = append(chain, func(in, out *value, l ...hs) *value {
				//fmt.Println("select index ", in, out)
				ctx.expr(s.Expr)
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
					fmt.Println(data)
					switch out.typ {
					case types.STRING:
						buf := []rune(out.toStr())
						fmt.Println(buf, i, buf[i])
						buf[i] = data.toRune()
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
	case ir.WrappedStatement:
		ctx.do(this.Fwd())
	case *ir.InvokeStmt:
		var par []interface{}
		for _, p := range this.Par {
			x := &param{}
			x.obj = p.Var
			if p.Expr != nil {
				ctx.expr(p.Expr)
				x.val = ctx.pop()
			} else {
				x.sel = p.Sel
			}
			par = append(par, x)
		}
		ctx.invoke(this.Mod, this.Proc, par...)
	case *ir.CallStmt:
		var par []interface{}
		for _, p := range this.Par {
			x := &param{}
			x.obj = p.Var
			if p.Expr != nil {
				ctx.expr(p.Expr)
				x.val = ctx.pop()
			} else {
				x.sel = p.Sel
			}
			par = append(par, x)
		}
		if this.Mod != "" {
			top := ctx.data.store[this.Mod]
			ctx.data.mpush(top.root)
		}
		ctx.do(this.Proc, par...)
		if this.Mod != "" {
			top := ctx.data.mpop()
			assert.For(top.Name == this.Mod, 60)
		}
	case *ir.AssignStmt:
		ctx.sel(this.Sel, nil, nil, func(in *value) *value {
			ctx.expr(this.Expr)
			val := ctx.pop()
			return val
		})
	case *ir.IfStmt:
		done := false
		for _, i := range this.Cond {
			ctx.expr(i.Expr)
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
				ctx.expr(i.Expr)
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
			ctx.expr(this.Cond.Expr)
			val := ctx.pop()
			stop = val.toBool()
		}
	case *ir.ChooseStmt:
		var base *ir.Dyadic
		if this.Expr != nil {
			base = &ir.Dyadic{}
			base.Op = operation.Eq
			base.Left = this.Expr
			//base.Right is open
		}
		done := false
		for _, i := range this.Cond {
			var ex ir.Expression
			if base != nil {
				base.Right = i.Expr
				ex = base
			} else {
				ex = i.Expr
			}
			assert.For(ex != nil, 40)
			ctx.expr(ex)
			val := ctx.pop()
			if val.toBool() {
				done = true
				for _, s := range i.Seq {
					ctx.do(s)
				}
				break
			} else if base != nil {
				base.Right = nil
			}
		}
		if !done && this.Else != nil {
			for _, s := range this.Else.Seq {
				ctx.do(s)
			}
		}
	default:
		halt.As(100, "unknown statement ", reflect.TypeOf(this))
	}
}
func (ctx *context) imp(i *ir.Import) {
	ms := ctx.data.store[i.Name]
	for _, x := range i.ConstDecl {
		c := x.This()
		mc := ms.root.ConstDecl[x.Name()]
		c.Expr = mc.Expr
	}
	for _, x := range i.VarDecl {
		v := x.This()
		mv := ms.root.VarDecl[x.Name()]
		v.Type = mv.Type
	}
	for _, x := range i.ProcDecl {
		p := x.This()
		mp := ms.root.ProcDecl[x.Name()]
		for k, v := range p.VarDecl {
			*v = *mp.VarDecl[k]
		}
		*p = *mp
	}
}

func (ctx *context) invoke(mod, proc string, par ...interface{}) (ret interface{}) {
	assert.For(rt.StdImp.Name == mod, 20)
	p := rt.StdImp.ProcDecl[proc]
	ctx.data.alloc(p)
	for _, _v := range par {
		switch v := _v.(type) {
		case *param:
			if v.val != nil {
				ctx.data.inner(v.obj, func(*value) *value { return v.val })
			} else {
				ctx.data.ref(v.obj, v.sel)
			}
		default:
			halt.As(100, "unknown par ", reflect.TypeOf(v))
		}
	}
	for i, e := range p.Pre() {
		ctx.expr(e)
		val := ctx.pop()
		assert.For(val.toBool(), 20+i)
	}
	if p := rt.StdProc[rt.Qualident{mod, proc}]; p != nil {
		p(ctx.data.top(), func(lt types.Type, l interface{}, op operation.Operation, rt types.Type, r interface{}, t types.Type) interface{} {
			rv := &value{typ: rt, val: r}
			lv := &value{typ: lt, val: l}
			v := calcDyadic(lv, op, rv)
			assert.For(v.typ == t, 60)
			return v.val
		})
	} else {
		halt.As(100, "unknown std procedure")
	}
	for i, e := range p.Post() {
		ctx.expr(e)
		val := ctx.pop()
		assert.For(val.toBool(), 60+i)
	}
	ret = ctx.data.dealloc(p)
	return
}

func (ctx *context) do(_t interface{}, par ...interface{}) (ret interface{}) {
	//	fmt.Println("do", reflect.TypeOf(_t))
	switch this := _t.(type) {
	case *ir.Module:
		for _, i := range this.ImportSeq {
			ctx.imp(i)
		}
		ctx.data.alloc(this)
		if len(this.BeginSeq) > 0 {
			for _, v := range this.BeginSeq {
				ctx.do(v)
			}
		}
		ctx.run()
		if len(this.CloseSeq) > 0 {
			for _, v := range this.CloseSeq {
				ctx.do(v)
			}
		}
		ctx.data.dealloc(this)
	case *ir.Procedure:
		ctx.data.alloc(this)
		for _, _v := range par {
			switch v := _v.(type) {
			case *param:
				if v.val != nil {
					ctx.data.inner(v.obj, func(*value) *value { return v.val })
				} else {
					ctx.data.ref(v.obj, v.sel)
				}
			default:
				halt.As(100, "unknown par ", reflect.TypeOf(v))
			}
		}
		for i, e := range this.Pre {
			ctx.expr(e)
			val := ctx.pop()
			assert.For(val.toBool(), 20+i)
		}
		for _, v := range this.Seq {
			ctx.do(v)
		}
		for i, e := range this.Post {
			ctx.expr(e)
			val := ctx.pop()
			assert.For(val.toBool(), 60+i)
		}
		ret = ctx.data.dealloc(this)
	case ir.Statement:
		ctx.stmt(this)
	default:
		halt.As(100, reflect.TypeOf(this))
	}
	return
}

func (c *context) run() {
	if len(c.load) > 0 {
		m := c.load[0]
		if len(c.load) > 1 {
			c.load = c.load[1:]
		} else {
			c.load = nil
		}
		c.do(m)
	}
}

func connectTo(m ...*ir.Module) (ret *context) {
	assert.For(len(m) > 0, 20)
	ret = &context{}
	ret.load = m
	ret.data = &storeStack{}
	ret.data.init(ret)
	ret.exprStack.init()
	return
}

func run(main *ir.Module, ld lenin.Loader) {

	cache := make(map[string]*ir.Module)
	var ml []string
	var do func(m *ir.Module)
	do = func(m *ir.Module) {

		ml = append(ml, m.Name)
		for _, v := range m.ImportSeq {
			assert.For(main.Name != v.Name, 30, "cyclic import from ", v.Name)
			if cache[v.Name] == nil {
				x, _ := ld(v.Name)
				assert.For(x != nil, 40)
				cache[v.Name] = x
				do(x)
			}
		}
	}
	cache[main.Name] = main
	do(main)
	var mm []*ir.Module
	for i := len(ml) - 1; i >= 0; i-- {
		if v := cache[ml[i]]; v != nil {
			mm = append(mm, v)
			delete(cache, ml[i])
		}
	}
	connectTo(mm...).run()
}

func init() {
	lenin.Run = run
}
