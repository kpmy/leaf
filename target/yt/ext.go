package yt

import (
	"github.com/kpmy/ypk/assert"
	"github.com/kpmy/ypk/halt"
	"leaf/ir"
	"reflect"
)

func externalize(mod *ir.Module) (ret *Module) {
	ret = &Module{}
	ret.init()
	ret.Name = mod.Name

	var expr func(ir.Expression) *Expression
	var sel func(s ir.Selector) []*Selector
	var stmt func(_s ir.Statement) (st *Statement)

	expr = func(_e ir.Expression) (ex *Expression) {
		ex = &Expression{}
		ex.Leaf = make(map[string]interface{})
		switch e := _e.(type) {
		case *ir.AtomExpr:
			ex.Type = Atom
			ex.Leaf["name"] = e.Value
		case *ir.ConstExpr:
			ex.Type = Constant
			ex.Leaf["value"] = e.Value
			ex.Leaf["type"] = e.Type.String()
		case *ir.NamedConstExpr:
			ex.Type = NamedConstant
			ex.Leaf["object"] = ret.this(e.Named)
		case *ir.VariableExpr:
			ex.Type = Variable
			ex.Leaf["object"] = ret.this(e.Obj)
		case *ir.Monadic:
			ex.Type = Monadic
			ex.Leaf["operand"] = expr(e.Operand)
			ex.Leaf["operation"] = e.Op.String()
		case *ir.Dyadic:
			ex.Type = Dyadic
			ex.Leaf["left"] = expr(e.Left)
			ex.Leaf["right"] = expr(e.Right)
			ex.Leaf["operation"] = e.Op.String()
		case *ir.SelectExpr:
			ex.Type = SelExpr
			ex.Leaf["base"] = expr(e.Base)
			ex.Leaf["selector"] = sel(e.Sel)
		case *dumbExpr:
			return expr(e.Eval())
		default:
			halt.As(100, "unexpected ", reflect.TypeOf(e))
		}
		return
	}
	sel = func(_s ir.Selector) (sl []*Selector) {
		x := &Selector{}
		x.Leaf = make(map[string]interface{})
		switch s := _s.(type) {
		case ir.ChainSelector:
			for _, v := range s.Chain() {
				tmp := sel(v)
				sl = append(sl, tmp...)
			}
		case *ir.SelectVar:
			x.Type = SelVar
			x.Leaf["object"] = ret.this(s.Var)
			sl = append(sl, x)
		case *ir.SelectIndex:
			x.Type = SelIdx
			x.Leaf["expression"] = expr(s.Expr.(ir.EvaluatedExpression).Eval())
			sl = append(sl, x)
		default:
			halt.As(100, "unknown selector ", reflect.TypeOf(s))
		}
		return
	}
	{
		for _, _v := range mod.ConstDecl {
			c := &Const{}
			c.Guid = ret.this(_v)
			var e ir.Expression
			switch v := _v.Expr.(type) {
			case ir.EvaluatedExpression:
				e = v.Eval()
			case *ir.AtomExpr:
				e = v
			default:
				halt.As(100, "unknown expression ", reflect.TypeOf(v))
			}
			assert.For(e != nil, 40)
			c.Expr = expr(e)
			ret.ConstDecl[_v.Name] = c
		}
	}
	{
		for _, v := range mod.VarDecl {
			i := &Var{}
			i.Guid = ret.this(v)
			i.Type = v.Type.String()
			ret.VarDecl[v.Name] = i
		}
	}
	stmt = func(_s ir.Statement) (st *Statement) {
		st = &Statement{}
		st.Leaf = make(map[string]interface{})
		switch s := _s.(type) {
		case *ir.AssignStmt:
			st.Type = Assign
			st.Leaf["selector"] = sel(s.Sel)
			e := s.Expr.(ir.EvaluatedExpression).Eval()
			st.Leaf["expression"] = expr(e)
		case *ir.IfStmt:
			st.Type = If
			var ifs []*Condition
			for _, v := range s.Cond {
				c := &Condition{}
				c.Expr = expr(v.Expr.(ir.EvaluatedExpression).Eval())
				for _, x := range v.Seq {
					c.Seq = append(c.Seq, stmt(x))
				}
				ifs = append(ifs, c)
			}
			st.Leaf["if"] = ifs
			if s.Else != nil {
				var ss []*Statement
				for _, x := range s.Else.Seq {
					ss = append(ss, stmt(x))
				}
				st.Leaf["else"] = ss
			}
		default:
			halt.As(100, "unexpected ", reflect.TypeOf(s))
		}
		return
	}
	{
		for _, v := range mod.BeginSeq {
			ret.BeginSeq = append(ret.BeginSeq, stmt(v))
		}
		for _, v := range mod.CloseSeq {
			ret.CloseSeq = append(ret.CloseSeq, stmt(v))
		}
	}
	return
}
