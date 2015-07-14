package lenin

import (
	"encoding/xml"
	"github.com/kpmy/ypk/halt"
	"leaf/ir/types"
)

type doraw struct {
	x *Any
}

func (r *doraw) Value() interface{} { return r.x }

func (r *doraw) MarshalXML(e *xml.Encoder, start xml.StartElement) (err error) {
	if start.Name.Local == "raw" {
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
