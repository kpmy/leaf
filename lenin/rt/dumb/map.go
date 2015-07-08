package dumb

import (
	"encoding/xml"
	"github.com/kpmy/ypk/halt"
	"leaf/ir/types"
	"leaf/lenin/trav"
)

type raw struct {
	x *trav.Any
}

func (r *raw) Convert() {}

func (r *raw) MarshalXML(e *xml.Encoder, start xml.StartElement) (err error) {
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
		m := x.(*trav.Map)
		for _, k := range m.Keys() {
			var pair xml.StartElement
			pair.Name.Local = "item"
			e.EncodeToken(pair)
			e.EncodeElement(&raw{x: k}, xml.StartElement{Name: xml.Name{Local: "key"}})
			v := m.Get(k)
			e.EncodeElement(&raw{x: v}, xml.StartElement{Name: xml.Name{Local: "value"}})
			e.EncodeToken(pair.End())
		}
	case types.LIST:
		l := x.(*trav.List)
		for i := 0; i < l.Len(); i++ {
			v := l.Get(i)
			e.EncodeElement(&raw{x: v}, xml.StartElement{Name: xml.Name{Local: "item"}})
		}
	case types.REAL:
		e.EncodeToken(xml.CharData([]byte(x.(*trav.Rat).String())))
	case types.STRING:
		e.EncodeToken(xml.CharData([]byte(x.(string))))
	default:
		halt.As(100, t, " ", x)
	}
	return e.EncodeToken(start.End())
}

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
