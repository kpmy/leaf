package lenin

import (
	"github.com/kpmy/ypk/halt"
	"leaf/ir/types"
	"leaf/lem"
	"reflect"
)

func Unvalue(_i interface{}) (ret *Any) {
	switch i := _i.(type) {
	case lem.Object:
		p := &Ptr{}
		x := i.Value()
		prepare(p, x.(*Any))
		ret = NewAny(types.PTR, p)
	case string:
		ret = NewAny(types.STRING, i)
	case map[interface{}]interface{}:
		ret = NewAny(types.MAP, Unmap(i))
	default:
		halt.As(100, reflect.TypeOf(i))
	}
	return
}

func Unmap(m map[interface{}]interface{}) (ret *Map) {
	ret = &Map{}
	for k, v := range m {
		kv := Unvalue(k)
		vv := Unvalue(v)
		ret.Set(kv, vv)
	}
	return
}
