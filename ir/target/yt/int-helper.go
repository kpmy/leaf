package yt

import (
	"github.com/kpmy/ypk/halt"
	"leaf/ir"
	"leaf/ir/modifiers"
	"leaf/ir/target/yt/fldz"
	"leaf/ir/types"
)

type ic struct {
	this *ir.Const
}

func (i *ic) Name() string { return i.this.Name }

func (i *ic) Expr() ir.Expression { return i.this.Expr }

func (i *ic) This() *ir.Const { return i.this }

type iv struct {
	this *ir.Variable
}

func (i *iv) Name() string { return i.this.Name }

func (i *iv) Type() types.Type { return i.this.Type }

func (i *iv) Modifier() modifiers.Modifier { return i.this.Modifier }

func (i *iv) This() *ir.Variable { return i.this }

type ip struct {
	this *ir.Procedure
}

func (i *ip) Name() string { return i.this.Name }

func (i *ip) VarDecl() map[string]ir.ImportVariable {
	vm := make(map[string]ir.ImportVariable)
	for k, v := range i.this.VarDecl {
		vm[k] = &iv{this: v}
	}
	return vm
}

func (i *ip) Infix() []ir.ImportVariable {
	var vl []ir.ImportVariable
	for _, v := range i.this.Infix {
		vl = append(vl, &iv{this: v})
	}
	return vl
}

func (i *ip) Pre() []ir.Expression  { return i.this.Pre }
func (i *ip) Post() []ir.Expression { return i.this.Post }
func (i *ip) This() *ir.Procedure   { return i.this }

type dumbExpr struct {
	e     ir.Expression
	later func() ir.Expression
}

func (d *dumbExpr) Self() {}
func (d *dumbExpr) Eval() ir.Expression {
	if d.e != nil {
		return d.e
	} else {
		return d.later()
	}
}

type dumbSel struct {
	chain []ir.Selector
}

func (d *dumbSel) Select() {}
func (d *dumbSel) Chain() []ir.Selector {
	return d.chain
}
func (d *dumbSel) put(s ir.Selector) { d.chain = append(d.chain, s) }

type dumbCall struct {
	c     *ir.CallStmt
	later func() *ir.CallStmt
}

func (d *dumbCall) Do() {}
func (d *dumbCall) Fwd() ir.Statement {
	if d.c != nil {
		return d.c
	} else {
		return d.later()
	}
}

func treatSel(_s interface{}) (ret *Selector) {
	ret = &Selector{}
	m := _s.(map[interface{}]interface{})
	ret.Type = SelType(m[fldz.Type].(string))
	ret.Leaf = make(map[string]interface{})
	leaf := m[fldz.Leaf].(map[interface{}]interface{})
	switch ret.Type {
	case SelVar:
		ret.Leaf[fldz.Object] = leaf[fldz.Object]
	case SelIdx:
		ret.Leaf[fldz.Expression] = leaf[fldz.Expression]
	case SelMod:
		ret.Leaf[fldz.Module] = leaf[fldz.Module]
	default:
		halt.As(100, "unexpected selector type ", ret.Type, " ", m)
	}
	return
}

func treatSelList(_l interface{}) (ret []*Selector) {
	l := _l.([]interface{})
	for _, s := range l {
		ret = append(ret, treatSel(s))
	}
	return
}

func treatExpr(_m interface{}) (ret *Expression) {
	ret = &Expression{}
	m := _m.(map[interface{}]interface{})
	ret.Type = ExprType(m[fldz.Type].(string))
	ret.Leaf = make(map[string]interface{})
	leaf := m[fldz.Leaf].(map[interface{}]interface{})
	switch ret.Type {
	case Constant:
		ret.Leaf[fldz.Value] = leaf[fldz.Value]
		ret.Leaf[fldz.Type] = leaf[fldz.Type]
	case NamedConstant:
		ret.Leaf[fldz.Object] = leaf[fldz.Object]
	case Set:
		ret.Leaf[fldz.Expression] = leaf[fldz.Expression]
	case Variable:
		ret.Leaf[fldz.Object] = leaf[fldz.Object]
	case TypeTest:
		ret.Leaf[fldz.Operand] = leaf[fldz.Operand]
		ret.Leaf[fldz.Type] = leaf[fldz.Type]
	case Monadic:
		ret.Leaf[fldz.Operand] = leaf[fldz.Operand]
		ret.Leaf[fldz.Operation] = leaf[fldz.Operation]
	case Dyadic:
		ret.Leaf[fldz.Left] = leaf[fldz.Left]
		ret.Leaf[fldz.Right] = leaf[fldz.Right]
		ret.Leaf[fldz.Operation] = leaf[fldz.Operation]
	case SelExpr:
		ret.Leaf[fldz.Base] = leaf[fldz.Base]
		ret.Leaf[fldz.Before] = leaf[fldz.Before]
		ret.Leaf[fldz.After] = leaf[fldz.After]
	case Infix, InvokeInfix:
		ret.Leaf[fldz.Procedure] = leaf[fldz.Procedure]
		ret.Leaf[fldz.Length] = leaf[fldz.Length]
		ret.Leaf[fldz.Operand] = leaf[fldz.Operand]
		ret.Leaf[fldz.Module] = leaf[fldz.Module]
	default:
		halt.As(100, "unexpected ", ret.Type, " ", _m)
	}
	return
}

func treatExprList(_l interface{}) (ret []*Expression) {
	l := _l.([]interface{})
	for _, c := range l {
		ret = append(ret, treatExpr(c))
	}
	return
}

func treatStmt(_m interface{}) (ret *Statement) {
	ret = &Statement{}
	m := _m.(map[interface{}]interface{})
	ret.Type = StmtType(m[fldz.Statement].(string))
	ret.Leaf = make(map[string]interface{})
	leaf := m[fldz.Leaf].(map[interface{}]interface{})
	switch ret.Type {
	case Call, Invoke:
		ret.Leaf[fldz.Parameter] = leaf[fldz.Parameter]
		ret.Leaf[fldz.Procedure] = leaf[fldz.Procedure]
		ret.Leaf[fldz.Module] = leaf[fldz.Module]
	case Assign:
		ret.Leaf[fldz.Selector] = leaf[fldz.Selector]
		ret.Leaf[fldz.Expression] = leaf[fldz.Expression]
	case If:
		ret.Leaf[fldz.Leaf] = leaf[fldz.Leaf]
		ret.Leaf[fldz.Else] = leaf[fldz.Else]
	case While:
		ret.Leaf[fldz.Leaf] = leaf[fldz.Leaf]
	case Repeat:
		ret.Leaf[fldz.Leaf] = leaf[fldz.Leaf]
	case Choose:
		ret.Leaf[fldz.Expression] = leaf[fldz.Expression]
		ret.Leaf[fldz.Leaf] = leaf[fldz.Leaf]
		ret.Leaf[fldz.Else] = leaf[fldz.Else]
		ret.Leaf[fldz.Type] = leaf[fldz.Type]
	default:
		halt.As(100, "unexpected ", ret.Type, " ", _m)
	}
	return
}

func treatBlock(_l interface{}) (ret []*Statement) {
	if _l != nil {
		l := _l.([]interface{})
		for _, s := range l {
			ret = append(ret, treatStmt(s))
		}
	}
	return
}

func treatIf(_m interface{}) (ret *Condition) {
	ret = &Condition{}
	m := _m.(map[interface{}]interface{})
	ret.Expr = treatExpr(m[fldz.Expression])
	ret.Seq = treatBlock(m[fldz.Block])
	return
}

func treatPar(_m interface{}) (ret *Param) {
	ret = &Param{}
	m := _m.(map[interface{}]interface{})
	if expr := m[fldz.Expression]; expr != nil {
		ret.Expr = treatExpr(expr)
	} else {
		ret.Sel = treatSelList(m[fldz.Selector])
	}
	ret.Uuid = m[fldz.UUID].(string)
	return
}

func treatParList(_l interface{}) (ret []*Param) {
	l := _l.([]interface{})
	for _, c := range l {
		ret = append(ret, treatPar(c))
	}
	return
}

func treatIfList(_l interface{}) (ret []*Condition) {
	l := _l.([]interface{})
	for _, c := range l {
		ret = append(ret, treatIf(c))
	}
	return
}

func treatElse(_l interface{}) (ret []*Statement) {
	ret = treatBlock(_l)
	return
}
