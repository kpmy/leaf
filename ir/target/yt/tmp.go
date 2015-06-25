package yt

import (
	"github.com/kpmy/ypk/halt"
	"leaf/ir"
	"leaf/ir/target/yt/fldz"
)

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
	case Variable:
		ret.Leaf[fldz.Object] = leaf[fldz.Object]
	case Monadic:
		ret.Leaf[fldz.Operand] = leaf[fldz.Operand]
		ret.Leaf[fldz.Operation] = leaf[fldz.Operation]
	case Dyadic:
		ret.Leaf["left"] = leaf["left"]
		ret.Leaf["right"] = leaf["right"]
		ret.Leaf["operation"] = leaf["operation"]
	case SelExpr:
		ret.Leaf["base"] = leaf["base"]
		ret.Leaf["selector"] = leaf["selector"]
	default:
		halt.As(100, "unexpected ", ret.Type, " ", _m)
	}
	return
}

func treatStmt(_m interface{}) (ret *Statement) {
	ret = &Statement{}
	m := _m.(map[interface{}]interface{})
	ret.Type = StmtType(m["statement"].(string))
	ret.Leaf = make(map[string]interface{})
	leaf := m["leaf"].(map[interface{}]interface{})
	switch ret.Type {
	case Call:
		ret.Leaf["param"] = leaf["param"]
		ret.Leaf["proc"] = leaf["proc"]
	case Assign:
		ret.Leaf["selector"] = leaf["selector"]
		ret.Leaf["expression"] = leaf["expression"]
	case If:
		ret.Leaf["leaf"] = leaf["leaf"]
		ret.Leaf["else"] = leaf["else"]
	case While:
		ret.Leaf["leaf"] = leaf["leaf"]
	case Repeat:
		ret.Leaf["leaf"] = leaf["leaf"]
	case Choose:
		ret.Leaf["expression"] = leaf["expression"]
		ret.Leaf["leaf"] = leaf["leaf"]
		ret.Leaf["else"] = leaf["else"]
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
	ret.Expr = treatExpr(m["expression"])
	ret.Seq = treatBlock(m["block"])
	return
}

func treatPar(_m interface{}) (ret *Param) {
	ret = &Param{}
	m := _m.(map[interface{}]interface{})
	ret.Expr = treatExpr(m["expression"])
	ret.Guid = m["guid"].(string)
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
