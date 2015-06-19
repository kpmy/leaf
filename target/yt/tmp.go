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
	default:
		halt.As(100, "unexpected ", _m, " ", ret.Type)
	}
	return
}
