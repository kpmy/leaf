package yt

import (
	"github.com/kpmy/ypk/assert"
	"github.com/kpmy/ypk/halt"
	"leaf/ir"
	"leaf/ir/target/yt/fldz"
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
			ex.Leaf[fldz.Name] = e.Value
		case *ir.ConstExpr:
			ex.Type = Constant
			ex.Leaf[fldz.Value] = e.Value
			ex.Leaf[fldz.Type] = e.Type.String()
		case *ir.NamedConstExpr:
			ex.Type = NamedConstant
			ex.Leaf[fldz.Object] = ret.this(e.Named)
		case *ir.VariableExpr:
			ex.Type = Variable
			ex.Leaf[fldz.Object] = ret.this(e.Obj)
		case *ir.Monadic:
			ex.Type = Monadic
			ex.Leaf[fldz.Operand] = expr(e.Operand)
			ex.Leaf[fldz.Operation] = e.Op.String()
		case *ir.Dyadic:
			ex.Type = Dyadic
			ex.Leaf[fldz.Left] = expr(e.Left)
			ex.Leaf[fldz.Right] = expr(e.Right)
			ex.Leaf[fldz.Operation] = e.Op.String()
		case *ir.SelectExpr:
			ex.Type = SelExpr
			ex.Leaf[fldz.Base] = expr(e.Base)
			ex.Leaf[fldz.Selector] = sel(e.Sel)
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
			x.Leaf[fldz.Object] = ret.this(s.Var)
			sl = append(sl, x)
		case *ir.SelectIndex:
			x.Type = SelIdx
			x.Leaf[fldz.Expression] = expr(s.Expr.(ir.EvaluatedExpression).Eval())
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
			st.Leaf[fldz.Procedure] = ret.this(s.Proc)
			var lp []*Param
			for _, p := range s.Par {
				par := &Param{}
				par.Uuid = ret.this(p.Var)
				if p.Expr != nil {
					par.Expr = expr(p.Expr.(ir.EvaluatedExpression).Eval())
				} else {
					par.Sel = sel(p.Sel)
				}
				lp = append(lp, par)
			}
			st.Leaf[fldz.Parameter] = lp
		case *ir.AssignStmt:
			st.Type = Assign
			st.Leaf[fldz.Selector] = sel(s.Sel)
			e := s.Expr.(ir.EvaluatedExpression).Eval()
			st.Leaf[fldz.Expression] = expr(e)
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
			st.Leaf[fldz.Leaf] = ifs
			if s.Else != nil {
				var ss []*Statement
				for _, x := range s.Else.Seq {
					ss = append(ss, stmt(x))
				}
				st.Leaf[fldz.Else] = ss
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
			st.Leaf[fldz.Leaf] = brs
		case *ir.RepeatStmt:
			st.Type = Repeat
			c := &Condition{}
			c.Expr = expr(s.Cond.Expr.(ir.EvaluatedExpression).Eval())
			for _, v := range s.Cond.Seq {
				c.Seq = append(c.Seq, stmt(v))
			}
			st.Leaf[fldz.Leaf] = c
		case *ir.ChooseStmt:
			st.Type = Choose
			if s.Expr != nil {
				st.Leaf[fldz.Expression] = expr(s.Expr.(ir.EvaluatedExpression).Eval())
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
			st.Leaf[fldz.Leaf] = brs
			if s.Else != nil {
				var ss []*Statement
				for _, x := range s.Else.Seq {
					ss = append(ss, stmt(x))
				}
				st.Leaf[fldz.Else] = ss
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
			c.Uuid = ret.this(_v)
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
			i.Uuid = ret.this(v)
			i.Type = v.Type.String()
			i.Modifier = v.Modifier.String()
			m[v.Name] = i
		}
		return
	}

	var pdecl func(pm map[string]*ir.Procedure) (m map[string]*Proc)
	pdecl = func(pm map[string]*ir.Procedure) (m map[string]*Proc) {
		m = make(map[string]*Proc)
		for _, v := range pm {
			i := &Proc{}
			i.Uuid = ret.this(v)
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
