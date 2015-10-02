package lenin

import (
	"fmt"
	"github.com/kpmy/leaf/ir"
	"github.com/kpmy/leaf/ir/operation"
	"github.com/kpmy/leaf/ir/types"
	"github.com/kpmy/leaf/lem"
	"github.com/kpmy/trigo"
	"github.com/kpmy/ypk/assert"
	"github.com/kpmy/ypk/fn"
	"github.com/kpmy/ypk/halt"
	"math/big"
	"reflect"
)

type context struct {
	data *storeStack
	exprStack
	begin  []*ir.Module
	end    []*ir.Module
	tgt    *targetStack
	mach   lem.Machine
	loader lem.Loader
	heap   *heap
	queue  []*later
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
			scope := ctx.tgt.pop()
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
			var a *Any
			switch v.typ {
			case types.ANY:
				a = v.toAny()
			case types.PTR:
				p := v.toPtr()
				if p.adr != 0 {
					a = p.link.Get()
				} else {
					a = &Any{}
				}
			default:
				halt.As(100, "unsupported ")
			}
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
		case *ir.SetExpr:
			var tmp []*value
			for _, x := range this.Expr {
				eval(x)
				v := ctx.pop()
				tmp = append(tmp, v)
			}
			ctx.push(&value{typ: types.SET, val: NewSet(tmp...)})
		case *ir.ListExpr:
			var tmp []*value
			for _, x := range this.Expr {
				eval(x)
				v := ctx.pop()
				tmp = append(tmp, v)
			}
			ctx.push(&value{typ: types.LIST, val: NewList(tmp...)})
		case *ir.MapExpr:
			var k []*value
			for _, x := range this.Key {
				eval(x)
				v := ctx.pop()
				k = append(k, v)
			}
			var v []*value
			for _, x := range this.Value {
				eval(x)
				n := ctx.pop()
				v = append(v, n)
			}
			ctx.push(&value{typ: types.MAP, val: NewMap(k, v)})
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
				case types.SET:
					s := v.toSet()
					ns := ThisSet(s)
					ns.inv = !ns.inv
					ctx.push(&value{typ: v.typ, val: ns})
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
				if this.Op == operation.In {
					l = &value{typ: types.ANY, val: ThisAny(l)}
				}
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
			if x, _ := ctx.do(this.Proc, pl...).(*storage); x != nil {
				out := this.Proc.Infix[0]
				val := x.data[out.Name]
				assert.For(val != nil, 40)
				ctx.push(&value{typ: out.Type, val: val.read()})
			} else {
				halt.As(100, "no result from infix")
			}
		case *ir.InvokeInfix:
			assert.For(lem.StdImp.Name == this.Mod, 20)
			proc := lem.StdImp.ProcDecl[this.Proc].This()
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
		case *ir.BindExpr:
			p := NewProc(this.Proc)
			ctx.push(&value{typ: types.PROC, val: p})
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
				ctx.tgt.push(ctx.data.store[s.Mod])
				//fmt.Println(s.Mod)
				return first(in, out, l...)
			})
		case *ir.SelectVar:
			chain = append(chain, func(in, out *value, l ...hs) (ret *value) {
				//fmt.Println("select var ", s.Var, in, out)
				scope := ctx.tgt.pop()
				ctx.data.outer(scope, s.Var, func(val *value) *value {
					ret = first(in, val, l...)
					return ret
				})
				return
			})
		case *ir.SelectInside:
			chain = append(chain, func(in, out *value, l ...hs) *value {
				//fmt.Println("select index ", in, out)
				ctx.expr(s.Expr)
				iv := ctx.pop()
				if in != nil { //get
					switch in.typ {
					case types.STRING:
						i := iv.toInt().Int64()
						buf := []rune(in.toStr())
						//fmt.Println(buf, i, buf[i])
						out = &value{typ: types.CHAR, val: buf[i]}
					case types.LIST:
						i := iv.toInt().Int64()
						l := in.asList()
						data := l.Get(int(i))
						out = &value{typ: types.ANY, val: data}
					case types.MAP:
						i := ThisAny(iv)
						m := in.asMap()
						data := m.Get(i)
						out = &value{typ: types.ANY, val: data}
					case types.PTR, types.ANY:
						switch iv.typ {
						case types.Undef: //dereference
							switch in.typ {
							case types.PTR:
								p := in.toPtr()
								if p.adr != 0 {
									data := p.link.Get()
									out = &value{typ: types.ANY, val: data}
								} else {
									halt.As(100, "nil dereference read")
								}
							case types.ANY:
								a := in.toAny()
								if !fn.IsNil(a.x) {
									out = &value{typ: a.typ, val: a.x}
								} else {
									halt.As(100, "undef dereference read")
								}
							default:
								halt.As(100, "cannot deref ", in.typ)
							}
						default:
							halt.As(100, "unknown selector type")
						}
					default:
						halt.As(100, "unknown base type ", in.typ)
					}
					return first(in, out, l...)
				} else if out != nil { //set
					data := first(in, out, l...)
					//fmt.Println(data)
					switch out.typ {
					case types.STRING:
						i := iv.toInt().Int64()
						buf := []rune(out.toStr())
						//fmt.Println(buf, i, buf[i])
						buf[i] = data.toRune()
						in = &value{typ: types.STRING, val: string(buf)}
					case types.LIST:
						i := iv.toInt().Int64()
						l := out.toList()
						l.Set(int(i), data)
						in = &value{typ: types.LIST, val: l}
					case types.MAP:
						i := ThisAny(iv)
						m := out.toMap()
						m.Set(i, ThisAny(data))
						in = &value{typ: types.MAP, val: m}
					case types.PTR, types.ANY:
						switch iv.typ {
						case types.Undef: //dereference
							switch out.typ {
							case types.PTR:
								p := out.toPtr()
								if p.adr != 0 {
									p.link.Set(ThisAny(data))
									in = &value{typ: types.PTR, val: ThisPtr(p)}
								} else {
									halt.As(100, "nil dereference write")
								}
							case types.ANY:
								a := out.toAny()
								if !fn.IsNil(a.x) {
									if data.typ == a.typ && data.typ != types.ANY {
										a.x = data.val
										out = &value{typ: a.typ, val: a.x}
									} else {
										halt.As(100, "incompatible types")
									}
								} else {
									halt.As(100, "undef dereference write")
								}
							default:
								halt.As(100, "cannot deref ", out.typ)
							}
						default:
							halt.As(100, "unknown selector type")
						}
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

func (ctx *context) _stmt(_s ir.Statement) {
	switch this := _s.(type) {
	case ir.WrappedStatement:
		ctx.do(this.Fwd())
	case *ir.InvokeStmt:
		var par []interface{}
		for _, p := range this.Par {
			x := &param{}
			x.obj = p.Var
			x.name = p.Variadic
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
		assert.For(len(this.Par) < 20, 20)
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
		ctx.do(this.Proc, par...)
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
		done := false
		if !this.TypeTest {
			var base *ir.Dyadic
			if this.Expr != nil {
				base = &ir.Dyadic{}
				base.Op = operation.Eq
				base.Left = this.Expr
				//base.Right is open
			}
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
		} else {
			e := this.Expr
			wrap := ""
			if ee, _ := e.(ir.EvaluatedExpression); ee != nil {
				e = ee.Eval()
			}
			switch s := e.(type) {
			case *ir.VariableExpr:
				//it's ok
				wrap = s.Obj.Name
			default:
				halt.As(100, "unsupported ", reflect.TypeOf(s))
			}
			ctx.expr(e)
			base := ctx.pop()
			for _, i := range this.Cond {
				var ex ir.Expression
				ex = i.Expr
				if ee, _ := ex.(ir.EvaluatedExpression); ee != nil {
					ex = ee.Eval()
				}
				//skip expression in this implementation
				switch t := ex.(type) {
				case *ir.TypeTest:
					x := base.toAny()
					if !fn.IsNil(x.x) && x.typ == t.Typ {
						done = true
						assert.For(wrap != "", 20)
						ctx.data.wrap(wrap, func(v *value) *value {
							av := v.toAny()
							nv := &value{typ: x.typ, val: av.x}
							return nv
						})
						for _, s := range i.Seq {
							ctx.do(s)
						}
						ctx.data.wrap(wrap, nil)
					}
				case *ir.Dyadic:
					//null expected
					x := base.toAny()
					if fn.IsNil(x.x) {
						done = true
						for _, s := range i.Seq {

							ctx.do(s)
						}
					}
				default:
					halt.As(100, reflect.TypeOf(t))
				}
				if done {
					break
				}
			}
		}
		if !done && this.Else != nil {
			for _, s := range this.Else.Seq {
				ctx.do(s)
			}
		} else if !done && this.TypeTest {
			halt.As(100, "NO ELSE")
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
	assert.For(lem.StdImp.Name == mod, 20)
	p := lem.StdImp.ProcDecl[proc]
	ctx.data.alloc(p)
	var varPar []lem.VarPar
	for _, _v := range par {
		switch v := _v.(type) {
		case *param:
			if v.obj != nil {
				if v.val != nil {
					ctx.data.inner(v.obj, func(*value) *value { return v.val })
				} else {
					ctx.data.ref(v.obj, v.sel)
				}
			} else {
				x := lem.VarPar{}
				x.Name = v.name
				x.Sel = v.sel
				x.Val = v.val
				varPar = append(varPar, x)
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
	if p := lem.StdProc[lem.Qualident{mod, proc}]; p != nil {
		p(ctx, ctx.data.top(), func(lt types.Type, l interface{}, op operation.Operation, rt types.Type, r interface{}, t types.Type) interface{} {
			rv := &value{typ: rt, val: r}
			lv := &value{typ: lt, val: l}
			v := calcDyadic(lv, op, rv)
			assert.For(v.typ == t, 60)
			return v.val
		}, varPar...)
	} else {
		halt.As(100, "unknown std procedure ", mod, ".", proc)
	}
	for i, e := range p.Post() {
		ctx.expr(e)
		val := ctx.pop()
		assert.For(val.toBool(), 60+i)
	}
	ret = ctx.data.dealloc(p)
	return
}

func (ctx *context) Mt(mt ...string) {
	if len(mt) > 0 {
		ctx.data.mt = mt[0]
		assert.For(ctx.data.store[mt[0]] != nil, 60, mt[0])
	} else {
		ctx.data.mt = ""
	}
}

func (ctx *context) ModuleOf(this *ir.Procedure) string {
	assert.For(this.Mod != "", 60, this, this.Name)
	return this.Mod
}

func (ctx *context) do(_t interface{}, par ...interface{}) (ret interface{}) {
	//	fmt.Println("do", reflect.TypeOf(_t))
	switch this := _t.(type) {
	case string: //dyn load, string invoke etc
		if nm, err := ctx.loader(this); err == nil {
			ml := importChain(nm, ctx.loader)
			for _, m := range ml {
				if ctx.data.store[m.Name] == nil {
					ctx.do(m)
				}
			}
		} else {
			halt.As(100, "error loading module ", this, " ", err)
		}
	case lem.Object: //broadcast
		for _, v := range ctx.data.store {
			am := this.Value().(*Any)
			assert.For(am.typ == types.MAP, 20)
			m := am.x.(*Map)
			if p := v.root.ProcDecl[lem.HANDLE]; p != nil {
				if v := p.VarDecl[lem.MSG]; v != nil {
					par := &param{}
					par.obj = v
					par.val = &value{typ: types.MAP, val: m}
					ctx.do(p, par)
				}
			}
		}
	case *ir.Module:
		for _, i := range this.ImportSeq {
			ctx.imp(i)
		}
		ctx.data.alloc(this)
		ctx.Mt(this.Name)
		ctx.end = append(ctx.end, this)
		if len(this.BeginSeq) > 0 {
			for _, v := range this.BeginSeq {
				ctx.do(v)
			}
		}
		ctx.start()
		ctx.Mt()
	case *ir.Procedure:
		//prologue
		old := ctx.data.mt
		ctx.Mt(ctx.ModuleOf(this))
		ctx.data.alloc(this)
		if lem.Debug {
			fmt.Println("PARAMS", len(par), fmt.Sprint(par...))
		}
		for _, _v := range par {
			switch v := _v.(type) {
			case *param:
				if v.val != nil {
					ctx.data.inner(v.obj, func(*value) *value { return v.val })
				} else {
					ctx.data.ref(v.obj, v.sel)
				}
			case lem.VarPar:
				obj := ctx.data.top().schema[v.Name]
				assert.For(obj != nil, 30)
				if val := v.Val.(*value); val != nil {
					ctx.data.inner(obj, func(*value) *value { return val })
				} else {
					ctx.data.ref(obj, v.Sel)
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
		//epilogue
		ctx.Mt(old)
	case ir.Statement:
		ctx._stmt(this)
		//очередь инструкций от системы
		for ctx.queue != nil {
			tmp := ctx.queue
			ctx.queue = nil
			for i := 0; i < len(tmp)-2; i++ {
				ctx.queue = append(ctx.queue, tmp[i])
			}
			this := tmp[len(tmp)-1]
			ctx.do(this.x, this.par...)
		}
	default:
		halt.As(100, reflect.TypeOf(this))
	}
	return
}

func (c *context) doLater(x interface{}, par ...interface{}) {
	l := &later{}
	l.x = x
	l.par = par
	tmp := c.queue
	c.queue = nil
	c.queue = append(c.queue, l)
	c.queue = append(c.queue, tmp...)
}

func (c *context) Queue(x interface{}, par ...lem.VarPar) {
	var p []interface{}
	for _, x := range par {
		p = append(p, x)
	}
	c.doLater(x, p...)
}

func (c *context) start() {
	if len(c.begin) > 0 {
		m := c.begin[0]
		if len(c.begin) > 1 {
			c.begin = c.begin[1:]
		} else {
			c.begin = nil
		}
		c.do(m)
	}
}

func (c *context) stop() {
	for i := len(c.end) - 1; i >= 0; i-- {
		this := c.end[i]
		c.Mt(this.Name)
		if len(this.CloseSeq) > 0 {
			for _, v := range this.CloseSeq {
				c.do(v)
			}
		}
		c.data.dealloc(this)
	}
}

func (c *context) Handler() func(lem.Message) lem.Message {
	return func(in lem.Message) lem.Message {
		c.mach.Chan() <- in
		return <-c.mach.Chan()
	}
}

func connectTo(mach lem.Machine, ld lem.Loader, m ...*ir.Module) (ret *context) {
	assert.For(len(m) > 0, 20)
	ret = &context{}
	ret.begin = m
	ret.mach = mach
	ret.data = &storeStack{}
	ret.tgt = &targetStack{}
	ret.loader = ld
	ret.data.init(ret)
	ret.tgt.init(ret)
	ret.heap = newHeap()
	ret.exprStack.init()
	return
}

func importChain(main *ir.Module, ld lem.Loader) []*ir.Module {
	assert.For(main != nil, 20)
	cache := make(map[string]*ir.Module)
	var ml []string
	var do func(m *ir.Module)
	do = func(m *ir.Module) {
		ml = append(ml, m.Name) // последовательность загрузки, модули могут повторяться и фильтруются ниже
		for _, v := range m.ImportSeq {
			assert.For(main.Name != v.Name, 30, "cyclic import from ", v.Name)
			if x := cache[v.Name]; x == nil {
				x, _ = ld(v.Name)
				assert.For(x != nil, 40, v.Name)
				cache[v.Name] = x
				do(x)
			} else {
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
	return mm
}

func _run(main *ir.Module, ld lem.Loader, m lem.Machine) {
	modList := importChain(main, ld)
	ctx := connectTo(m, ld, modList...)
	msg := map[interface{}]interface{}{"type": "machine", "context": ctx}
	m.Chan() <- msg
	<-m.Chan()
	ctx.start()
	ctx.stop()
}
