package lenin

import (
	"encoding/xml"
	"github.com/kpmy/leaf/ir/types"
	"github.com/kpmy/leaf/lem"
	"github.com/kpmy/ypk/halt"
	"reflect"
)

type doraw struct {
	x *Any
}

func (r *doraw) Value() interface{} { return r.x }

func (r *doraw) MarshalXML(e *xml.Encoder, start xml.StartElement) (err error) {
	if start.Name.Local == "doraw" {
		start.Name.Local = "any"
	}
	t, x := r.x.This()
	var ta xml.Attr
	ta.Name.Local = "type"
	ta.Value = t.String()
	start.Attr = append(start.Attr, ta)
	e.EncodeToken(start)
	switch t {
	case types.PTR:
		halt.As(100, "no pointers inside")
	case types.MAP:
		m := x.(*Map)
		for _, k := range m.Keys() {
			var pair xml.StartElement
			pair.Name.Local = "item"
			e.EncodeToken(pair)
			e.EncodeElement(&doraw{x: k}, xml.StartElement{Name: xml.Name{Local: "key"}})
			v := m.Get(k)
			e.EncodeElement(&doraw{x: v}, xml.StartElement{Name: xml.Name{Local: "value"}})
			e.EncodeToken(pair.End())
		}
	case types.LIST:
		l := x.(*List)
		for i := 0; i < l.Len(); i++ {
			v := l.Get(i)
			e.EncodeElement(&doraw{x: v}, xml.StartElement{Name: xml.Name{Local: "item"}})
		}
	case types.REAL:
		e.EncodeToken(xml.CharData([]byte(x.(*Rat).String())))
	case types.STRING:
		e.EncodeToken(xml.CharData([]byte(x.(string))))
	default:
		halt.As(100, t, " ", x)
	}
	return e.EncodeToken(start.End())
}

func DoRaw(x *Any) (ret interface{}) {
	return &doraw{x: x}
}

func DoValue(_x *Any) (ret interface{}) {
	t, x := _x.This()
	switch t {
	case types.ATOM:
		ret = string(x.(Atom))
	case types.STRING:
		ret = x.(string)
	case types.BOOLEAN:
		ret = x.(bool)
	case types.INTEGER:
		ret = x.(*Int).Int64()
	case types.PTR: //pointer goes raw
		ptr := x.(*Ptr)
		ret = DoRaw(ptr.Get())
	case types.MAP:
		ret = DoMap(x.(*Map))
	case types.LIST:
		ret = DoList(x.(*List))
	case types.REAL:
		ret, _ = x.(*Rat).Float64()
	default:
		halt.As(100, x, " ", t)
	}
	return
}

func DoList(l *List) (ret []interface{}) {
	for i := 0; i < l.Len(); i++ {
		ret = append(ret, DoValue(l.Get(i)))
	}
	return
}

func DoMap(m *Map) (ret map[interface{}]interface{}) {
	ret = make(map[interface{}]interface{})
	for _, k := range m.Keys() {
		ret[DoValue(k)] = DoValue(m.Get(k))
	}
	return
}

func Unvalue(h *heap, _i interface{}) (ret *Any) {
	switch i := _i.(type) {
	case lem.Object:
		p := &Ptr{}
		x := i.Value()
		h.prepare(p, x.(*Any))
		ret = NewAny(types.PTR, p)
	case string:
		ret = NewAny(types.STRING, i)
	case map[interface{}]interface{}:
		ret = NewAny(types.MAP, Unmap(h, i))
	case lem.Message:
		ret = NewAny(types.MAP, Unmap(h, i))
	default:
		halt.As(100, reflect.TypeOf(i))
	}
	return
}

func Unmap(h *heap, m map[interface{}]interface{}) (ret *Map) {
	ret = &Map{}
	for k, v := range m {
		kv := Unvalue(h, k)
		vv := Unvalue(h, v)
		ret.Set(kv, vv)
	}
	return
}
