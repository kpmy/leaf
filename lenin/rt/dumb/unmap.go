package dumb

import (
	"github.com/kpmy/ypk/halt"
	"leaf/ir/types"
	"leaf/lem"
	"leaf/lenin/trav"
	"reflect"
)

func Unvalue(_i interface{}) (ret *trav.Any) {
	switch i := _i.(type) {
	case lem.Object:
		p := &trav.Ptr{}
		x := i.Value()
		prepare(p, x)
		ret = trav.NewAny(types.PTR, p)
	case string:
		ret = trav.NewAny(types.STRING, i)
	default:
		halt.As(100, reflect.TypeOf(i))
	}
	return
}

func Unmap(m map[interface{}]interface{}) (ret *trav.Map) {
	ret = &trav.Map{}
	for k, v := range m {
		kv := Unvalue(k)
		vv := Unvalue(v)
		ret.Set(kv, vv)
	}
	return
}
