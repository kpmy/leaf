package yt

import (
	"github.com/kpmy/ypk/assert"
	"github.com/kpmy/ypk/halt"
	"leaf/ir"
	"leaf/ir/modifiers"
	"leaf/ir/operation"
	"leaf/ir/target/yt/fldz"
	"leaf/ir/types"
)

func internalize(m *Module) (ret *ir.Module) {
	ret = &ir.Module{}
	ret.Init()
	ret.Name = m.Name
	var expr func(e *Expression) ir.Expression
	var sel func(ls []*Selector) ir.Selector
	var stmt func(s *Statement) (ret ir.Statement)

	expr = func(e *Expression) ir.Expression {
		d := &dumbExpr{}
		switch e.Type {
		case Atom:
			this := &ir.AtomExpr{}
			this.Value = e.Leaf[fldz.Name].(string)
			d.later = func() ir.Expression { return this }
		case Constant:
			this := &ir.ConstExpr{}
			this.Value = e.Leaf[fldz.Value]
			this.Type = types.TypMap[e.Leaf[fldz.Type].(string)]
			typeFix(this)
			d.e = this
		case NamedConstant:
			this := &ir.NamedConstExpr{}
			id := e.Leaf[fldz.Object].(string)
			_n := m.that(id)
			if n, ok := _n.(*ir.Const); ok {
				this.Named = n
				d.e = this
			} else if _n != nil {
				d.later = func() ir.Expression {
					fn := _n.(func() interface{})
					if x, ok := fn().(*ir.Const); ok {
						this.Named = x
						return this
					} else {
						halt.As(101, "wrong future expr")
					}
					panic(0)
				}
			} else {
				halt.As(100, "unexpected nil")
			}
		case Variable:
			this := &ir.VariableExpr{}
			id := e.Leaf[fldz.Object].(string)
			this.Obj = m.that(id).(*ir.Variable)
			d.e = this
		case Monadic:
			this := &ir.Monadic{}
			this.Operand = expr(treatExpr(e.Leaf[fldz.Operand]))
			this.Op = operation.OpMap[e.Leaf[fldz.Operation].(string)]
			d.e = this
		case Dyadic:
			this := &ir.Dyadic{}
			this.Left = expr(treatExpr(e.Leaf[fldz.Left]))
			this.Right = expr(treatExpr(e.Leaf[fldz.Right]))
			this.Op = operation.OpMap[e.Leaf[fldz.Operation].(string)]
			d.e = this
		case SelExpr:
			this := &ir.SelectExpr{}
			this.Base = expr(treatExpr(e.Leaf[fldz.Base]))
			this.Sel = sel(treatSelList(e.Leaf[fldz.Selector]))
			d.e = this
		default:
			halt.As(100, "unknown type ", e.Type)
		}
		assert.For(d.e != nil || d.later != nil, 60)
		return d
	}
	sel = func(ls []*Selector) ir.Selector {
		ds := &dumbSel{}
		for _, s := range ls {
			switch s.Type {
			case SelVar:
				this := &ir.SelectVar{}
				id := s.Leaf[fldz.Object].(string)
				this.Var = m.that(id).(*ir.Variable)
				ds.put(this)
			case SelIdx:
				this := &ir.SelectIndex{}
				this.Expr = expr(treatExpr(s.Leaf[fldz.Expression]))
				ds.put(this)
			default:
				halt.As(100, "unknown type ", s.Type)
			}
		}
		assert.For(len(ds.chain) > 0, 60)
		return ds
	}
	stmt = func(s *Statement) (ret ir.Statement) {
		switch s.Type {
		case Call:
			d := &dumbCall{}
			this := &ir.CallStmt{}
			_n := m.that(s.Leaf[fldz.Procedure].(string))
			pl := treatParList(s.Leaf[fldz.Parameter])
			if n, ok := _n.(*ir.Procedure); ok {
				this.Proc = n
				for _, par := range pl {
					x := &ir.Parameter{}
					if par.Expr != nil {
						x.Expr = expr(par.Expr)
					} else {
						x.Sel = sel(par.Sel)
					}
					x.Var = m.that(par.Uuid).(*ir.Variable)
					this.Par = append(this.Par, x)
				}
				d.c = this
			} else if _n != nil {
				d.later = func() *ir.CallStmt {
					fn := _n.(func() interface{})
					if x, ok := fn().(*ir.Procedure); ok {
						this.Proc = x
						for _, par := range pl {
							x := &ir.Parameter{}
							x.Expr = expr(par.Expr)
							x.Var = m.that(par.Uuid).(*ir.Variable)
							this.Par = append(this.Par, x)
						}
						return this
					} else {
						halt.As(101, "wrong forward call")
					}
					panic(0)
				}
			} else {
				halt.As(100, "unexpected nil")
			}
			ret = d
		case Assign:
			this := &ir.AssignStmt{}
			this.Sel = sel(treatSelList(s.Leaf[fldz.Selector]))
			this.Expr = expr(treatExpr(s.Leaf[fldz.Expression]))
			ret = this
		case If:
			this := &ir.IfStmt{}
			cl := treatIfList(s.Leaf[fldz.Leaf])
			sl := treatElse(s.Leaf[fldz.Else])
			for _, c := range cl {
				i := &ir.ConditionBranch{}
				i.Expr = expr(c.Expr)
				for _, s := range c.Seq {
					i.Seq = append(i.Seq, stmt(s))
				}
				this.Cond = append(this.Cond, i)
			}
			if sl != nil {
				e := &ir.ElseBranch{}
				for _, s := range sl {
					e.Seq = append(e.Seq, stmt(s))
				}
				this.Else = e
			}
			ret = this
		case While:
			this := &ir.WhileStmt{}
			cl := treatIfList(s.Leaf[fldz.Leaf])
			for _, c := range cl {
				i := &ir.ConditionBranch{}
				i.Expr = expr(c.Expr)
				for _, s := range c.Seq {
					i.Seq = append(i.Seq, stmt(s))
				}
				this.Cond = append(this.Cond, i)
			}
			ret = this
		case Repeat:
			this := &ir.RepeatStmt{}
			c := treatIf(s.Leaf[fldz.Leaf])
			i := &ir.ConditionBranch{}
			i.Expr = expr(c.Expr)
			for _, s := range c.Seq {
				i.Seq = append(i.Seq, stmt(s))
			}
			this.Cond = i
			ret = this
		case Choose:
			this := &ir.ChooseStmt{}
			cl := treatIfList(s.Leaf[fldz.Leaf])
			for _, c := range cl {
				i := &ir.ConditionBranch{}
				i.Expr = expr(c.Expr)
				for _, s := range c.Seq {
					i.Seq = append(i.Seq, stmt(s))
				}
				this.Cond = append(this.Cond, i)
			}
			if s.Leaf[fldz.Expression] != nil {
				this.Expr = expr(treatExpr(s.Leaf[fldz.Expression]))
			}
			sl := treatElse(s.Leaf[fldz.Else])
			if sl != nil {
				e := &ir.ElseBranch{}
				for _, s := range sl {
					e.Seq = append(e.Seq, stmt(s))
				}
				this.Else = e
			}
			ret = this
		default:
			halt.As(100, "unexpected ", s.Type)
		}
		return
	}
	cdecl := func(cm map[string]*Const) (im map[string]*ir.Const) {
		im = make(map[string]*ir.Const)
		for k, v := range cm {
			c := &ir.Const{}
			c.Name = k
			c.Expr = expr(v.Expr)
			m.that(v.Uuid, c)
			im[k] = c
		}
		return
	}
	vdecl := func(vm map[string]*Var) (im map[string]*ir.Variable) {
		im = make(map[string]*ir.Variable)
		for k, v := range vm {
			i := &ir.Variable{}
			i.Name = k
			i.Type = types.TypMap[v.Type]
			i.Modifier = modifiers.ModMap[v.Modifier]
			m.that(v.Uuid, i)
			im[k] = i
		}
		return
	}
	var pdecl func(pm map[string]*Proc) (im map[string]*ir.Procedure)
	pdecl = func(pm map[string]*Proc) (im map[string]*ir.Procedure) {
		im = make(map[string]*ir.Procedure)
		for k, v := range pm {
			p := &ir.Procedure{}
			p.Name = k
			p.ConstDecl = cdecl(v.ConstDecl)
			p.VarDecl = vdecl(v.VarDecl)
			p.ProcDecl = pdecl(v.ProcDecl)
			for _, s := range v.Seq {
				p.Seq = append(p.Seq, stmt(s))
			}
			m.that(v.Uuid, p)
			im[k] = p
		}
		return
	}
	{
		ret.ConstDecl = cdecl(m.ConstDecl)
		ret.VarDecl = vdecl(m.VarDecl)
		ret.ProcDecl = pdecl(m.ProcDecl)
		for _, v := range m.BeginSeq {
			ret.BeginSeq = append(ret.BeginSeq, stmt(v))
		}
		for _, v := range m.CloseSeq {
			ret.CloseSeq = append(ret.CloseSeq, stmt(v))
		}
	}
	return
}
