package dumb

import (
	"github.com/kpmy/ypk/halt"
	"leaf/ir/types"
	"leaf/lenin/trav"
)

func Value(_x *trav.Any) (ret interface{}) {
	t, x := _x.This()
	switch t {
	case types.ATOM:
		ret = string(x.(trav.Atom))
	case types.STRING:
		ret = x.(string)
	case types.BOOLEAN:
		ret = x.(bool)
	case types.INTEGER:
		ret = x.(*trav.Int).String()
	default:
		halt.As(100, x, " ", t)
	}
	return
}

func Map(m *trav.Map) (ret map[interface{}]interface{}) {
	ret = make(map[interface{}]interface{})
	for _, k := range m.Keys() {
		ret[Value(k)] = Value(m.Get(k))
	}
	return
}
