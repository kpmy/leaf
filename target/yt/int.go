package yt

import (
	"github.com/kpmy/ypk/assert"
	"github.com/kpmy/ypk/halt"
	"leaf/ir"
	"leaf/ir/operation"
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
			this.Value = e.Leaf["name"].(string)
			d.later = func() ir.Expression { return this }
		case Constant:
			this := &ir.ConstExpr{}
			this.Value = e.Leaf["value"]
			this.Type = types.TypMap[e.Leaf["type"].(string)]
			typeFix(this)
			d.e = this
		case NamedConstant:
			this := &ir.NamedConstExpr{}
			id := e.Leaf["object"].(string)
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
			id := e.Leaf["object"].(string)
			this.Obj = m.that(id).(*ir.Variable)
			d.e = this
		case Monadic:
			this := &ir.Monadic{}
			this.Operand = expr(treatExpr(e.Leaf["operand"]))
			this.Op = operation.OpMap[e.Leaf["operation"].(string)]
			d.e = this
		case Dyadic:
			this := &ir.Dyadic{}
			this.Left = expr(treatExpr(e.Leaf["left"]))
			this.Right = expr(treatExpr(e.Leaf["right"]))
			this.Op = operation.OpMap[e.Leaf["operation"].(string)]
			d.e = this
		case SelExpr:
			this := &ir.SelectExpr{}
			this.Base = expr(treatExpr(e.Leaf["base"]))
			this.Sel = sel(treatSelList(e.Leaf["selector"]))
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
				id := s.Leaf["object"].(string)
				this.Var = m.that(id).(*ir.Variable)
				ds.put(this)
			case SelIdx:
				this := &ir.SelectIndex{}
				this.Expr = expr(treatExpr(s.Leaf["expression"]))
				ds.put(this)
			default:
				halt.As(100, "unknown type ", s.Type)
			}
		}
		assert.For(len(ds.chain) > 0, 60)
		return ds
	}
	{
		for k, v := range m.ConstDecl {
			c := &ir.Const{}
			c.Name = k
			c.Expr = expr(v.Expr)
			m.that(v.Guid, c)
			ret.ConstDecl[k] = c
		}
	}

	{
		for k, v := range m.VarDecl {
			i := &ir.Variable{}
			i.Name = k
			i.Type = types.TypMap[v.Type]
			m.that(v.Guid, i)
			ret.VarDecl[k] = i
		}
	}
	stmt = func(s *Statement) (ret ir.Statement) {
		switch s.Type {
		case Assign:
			this := &ir.AssignStmt{}
			this.Sel = sel(treatSelList(s.Leaf["selector"]))
			this.Expr = expr(treatExpr(s.Leaf["expression"]))
			ret = this
		case If:
			this := &ir.IfStmt{}
			cl := treatIfList(s.Leaf["if"])
			sl := treatElse(s.Leaf["else"])
			for _, c := range cl {
				i := &ir.IfBranch{}
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
		default:
			halt.As(100, "unexpected ", s.Type)
		}
		return
	}
	{
		for _, v := range m.BeginSeq {
			ret.BeginSeq = append(ret.BeginSeq, stmt(v))
		}
		for _, v := range m.CloseSeq {
			ret.CloseSeq = append(ret.CloseSeq, stmt(v))
		}
	}
	return
}
