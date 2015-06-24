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
	stmt = func(s *Statement) (ret ir.Statement) {
		switch s.Type {
		case Call:
			d := &dumbCall{}
			this := &ir.CallStmt{}
			_n := m.that(s.Leaf["proc"].(string))
			if n, ok := _n.(*ir.Procedure); ok {
				this.Proc = n
				d.c = this
			} else if _n != nil {
				d.later = func() *ir.CallStmt {
					fn := _n.(func() interface{})
					if x, ok := fn().(*ir.Procedure); ok {
						this.Proc = x
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
			this.Sel = sel(treatSelList(s.Leaf["selector"]))
			this.Expr = expr(treatExpr(s.Leaf["expression"]))
			ret = this
		case If:
			this := &ir.IfStmt{}
			cl := treatIfList(s.Leaf["leaf"])
			sl := treatElse(s.Leaf["else"])
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
			cl := treatIfList(s.Leaf["leaf"])
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
			c := treatIf(s.Leaf["leaf"])
			i := &ir.ConditionBranch{}
			i.Expr = expr(c.Expr)
			for _, s := range c.Seq {
				i.Seq = append(i.Seq, stmt(s))
			}
			this.Cond = i
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
			m.that(v.Guid, c)
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
			m.that(v.Guid, i)
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
			m.that(v.Guid, p)
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
