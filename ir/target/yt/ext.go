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
	stmt = func(_s ir.Statement) (st *Statement) {
		st = &Statement{}
		st.Leaf = make(map[string]interface{})
		switch s := _s.(type) {
		case ir.WrappedStatement:
			return stmt(s.Fwd())
		case *ir.CallStmt:
			st.Type = Call
			st.Leaf["proc"] = ret.this(s.Proc)
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
			st.Leaf["leaf"] = ifs
			if s.Else != nil {
				var ss []*Statement
				for _, x := range s.Else.Seq {
					ss = append(ss, stmt(x))
				}
				st.Leaf["else"] = ss
			}
		case *ir.WhileStmt:
			st.Type = While
			var brs []*Condition
			for _, v := range s.Cond {
				c := &Condition{}
				c.Expr = expr(v.Expr.(ir.EvaluatedExpression).Eval())
				for _, x := range v.Seq {
					c.Seq = append(c.Seq, stmt(x))
				}
				brs = append(brs, c)
			}
			st.Leaf["leaf"] = brs
		case *ir.RepeatStmt:
			st.Type = Repeat
			c := &Condition{}
			c.Expr = expr(s.Cond.Expr.(ir.EvaluatedExpression).Eval())
			for _, v := range s.Cond.Seq {
				c.Seq = append(c.Seq, stmt(v))
			}
			st.Leaf["leaf"] = c
		case *ir.ChooseStmt:
			st.Type = Choose
			if s.Expr != nil {
				st.Leaf["expression"] = expr(s.Expr.(ir.EvaluatedExpression).Eval())
			}
			var brs []*Condition
			for _, v := range s.Cond {
				c := &Condition{}
				c.Expr = expr(v.Expr.(ir.EvaluatedExpression).Eval())
				for _, x := range v.Seq {
					c.Seq = append(c.Seq, stmt(x))
				}
				brs = append(brs, c)
			}
			st.Leaf["leaf"] = brs
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
	cdecl := func(cm map[string]*ir.Const) (m map[string]*Const) {
		m = make(map[string]*Const)
		for _, _v := range cm {
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
			m[_v.Name] = c
		}
		return
	}
	vdecl := func(vm map[string]*ir.Variable) (m map[string]*Var) {
		m = make(map[string]*Var)
		for _, v := range vm {
			i := &Var{}
			i.Guid = ret.this(v)
			i.Type = v.Type.String()
			m[v.Name] = i
		}
		return
	}

	var pdecl func(pm map[string]*ir.Procedure) (m map[string]*Proc)
	pdecl = func(pm map[string]*ir.Procedure) (m map[string]*Proc) {
		m = make(map[string]*Proc)
		for _, v := range pm {
			i := &Proc{}
			i.Guid = ret.this(v)
			i.ConstDecl = cdecl(v.ConstDecl)
			i.VarDecl = vdecl(v.VarDecl)
			i.ProcDecl = pdecl(v.ProcDecl)
			for _, s := range v.Seq {
				i.Seq = append(i.Seq, stmt(s))
			}
			m[v.Name] = i
		}
		return
	}
	{
		ret.ConstDecl = cdecl(mod.ConstDecl)
		ret.VarDecl = vdecl(mod.VarDecl)
		ret.ProcDecl = pdecl(mod.ProcDecl)
		for _, v := range mod.BeginSeq {
			ret.BeginSeq = append(ret.BeginSeq, stmt(v))
		}
		for _, v := range mod.CloseSeq {
			ret.CloseSeq = append(ret.CloseSeq, stmt(v))
		}
	}
	return
}
