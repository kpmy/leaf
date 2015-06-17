package yt

import (
	"github.com/kpmy/ypk/halt"
	"leaf/ir"
)

type dumbExpr struct {
	e ir.Expression
}

func (d *dumbExpr) Self() {}
func (d *dumbExpr) Eval() ir.Expression {
	return d.e
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
	case NamedConstant:
		ret.Leaf["object"] = leaf["object"]
	case Variable:
		ret.Leaf["object"] = leaf["object"]
	case Monadic:
		ret.Leaf["operand"] = leaf["operand"]
	case Dyadic:
		ret.Leaf["left"] = leaf["left"]
		ret.Leaf["right"] = leaf["right"]
	default:
		halt.As(100, "unexpected ", _m, " ", ret.Type)
	}
	return
}
