package yt

import (
	"github.com/kpmy/ypk/halt"
	"leaf/ir"
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

func treatSel(_s interface{}) (ret *Selector) {
	ret = &Selector{}
	m := _s.(map[interface{}]interface{})
	ret.Type = SelType(m["type"].(string))
	ret.Leaf = make(map[string]interface{})
	leaf := m["leaf"].(map[interface{}]interface{})
	switch ret.Type {
	case SelVar:
		ret.Leaf["object"] = leaf["object"]
	case SelIdx:
		ret.Leaf["expression"] = leaf["expression"]
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
	ret.Type = ExprType(m["type"].(string))
	ret.Leaf = make(map[string]interface{})
	leaf := m["leaf"].(map[interface{}]interface{})
	switch ret.Type {
	case Constant:
		ret.Leaf["value"] = leaf["value"]
		ret.Leaf["type"] = leaf["type"]
	case NamedConstant:
		ret.Leaf["object"] = leaf["object"]
	case Variable:
		ret.Leaf["object"] = leaf["object"]
	case Monadic:
		ret.Leaf["operand"] = leaf["operand"]
		ret.Leaf["operation"] = leaf["operation"]
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
