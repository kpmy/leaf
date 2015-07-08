package dumb

import (
	"github.com/kpmy/ypk/halt"
	"leaf/ir/types"
	"leaf/lenin/trav"
)

type raw struct {
	x *trav.Any
}

func (r *raw) Convert() {}

func Raw(x *trav.Any) (ret interface{}) {
	return &raw{x: x}
}

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
		ret = x.(*trav.Int).Int64()
	case types.PTR: //pointer goes raw
		ptr := x.(*trav.Ptr)
		ret = Raw(ptr.Get())
	case types.MAP:
		ret = Map(x.(*trav.Map))
	case types.LIST:
		ret = List(x.(*trav.List))
	case types.REAL:
		ret, _ = x.(*trav.Rat).Float64()
	default:
		halt.As(100, x, " ", t)
	}
	return
}

func List(l *trav.List) (ret []interface{}) {
	for i := 0; i < l.Len(); i++ {
		ret = append(ret, Value(l.Get(i)))
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
