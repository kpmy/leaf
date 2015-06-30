package trav

import (
	"github.com/kpmy/trigo"
	"github.com/kpmy/ypk/assert"
	"github.com/kpmy/ypk/halt"
	"leaf/ir"
	"leaf/ir/types"
	"math/big"
	"reflect"
)

type param struct {
	obj *ir.Variable
	val *value
	sel ir.Selector
}

type value struct {
	typ types.Type
	val interface{}
}

func (v *value) toInt() (ret *big.Int) {
	assert.For(v.typ == types.INTEGER, 20)
	switch x := v.val.(type) {
	case int:
		ret = big.NewInt(int64(x))
	case *Int:
		ret = big.NewInt(0)
		ret.Add(ret, &x.Int)
	default:
		halt.As(100, "wrong integer ", reflect.TypeOf(x))
	}
	return
}

func (v *value) toReal() (ret *big.Rat) {
	assert.For(v.typ == types.REAL, 20)
	switch x := v.val.(type) {
	case *Rat:
		ret = big.NewRat(0, 1)
		ret.Add(ret, &x.Rat)
	default:
		halt.As(100, "wrong real ", reflect.TypeOf(x))
	}
	return
}

func (v *value) toBool() (ret bool) {
	assert.For(v.typ == types.BOOLEAN, 20)
	switch x := v.val.(type) {
	case bool:
		ret = x
	default:
		halt.As(100, "wrong boolean ", reflect.TypeOf(x))
	}
	return
}

func (v *value) toStr() (ret string) {
	assert.For(v.typ == types.STRING, 20)
	switch x := v.val.(type) {
	case string:
		ret = x
	default:
		halt.As(100, "wrong string ", reflect.TypeOf(x))
	}
	return
}

func (v *value) toRune() (ret rune) {
	assert.For(v.typ == types.CHAR, 20, v.typ)
	switch x := v.val.(type) {
	case rune:
		ret = x
	default:
		halt.As(100, "wrong rune ", reflect.TypeOf(x))
	}
	return
}

func (v *value) toTril() (ret tri.Trit) {
	assert.For(v.typ == types.TRILEAN, 20)
	switch x := v.val.(type) {
	case tri.Trit:
		ret = x
	default:
		halt.As(100, "wrong trilean ", reflect.TypeOf(x))
	}
	return
}

func (v *value) toAtom() (ret Atom) {
	assert.For(v.typ == types.ATOM, 20)
	switch x := v.val.(type) {
	case Atom:
		ret = x
	case nil: //do nothing
	default:
		halt.As(100, "wrong atom ", reflect.TypeOf(x))
	}
	return
}

func (v *value) toCmp() (ret *Cmp) {
	assert.For(v.typ == types.COMPLEX, 20)
	switch x := v.val.(type) {
	case *Cmp:
		ret = ThisCmp(x)
	default:
		halt.As(100, "wrong complex ", reflect.TypeOf(x))
	}
	return
}

func (v *value) toAny() (ret *Any) {
	assert.For(v.typ == types.ANY, 20)
	switch x := v.val.(type) {
	case *Any:
		ret = ThisAny(&value{typ: x.typ, val: x.x})
	default:
		halt.As(100, "wrong any ", reflect.TypeOf(x))
	}
	return
}

func cval(e *ir.ConstExpr) (ret *value) {
	t := e.Type
	switch t {
	case types.INTEGER:
		b := big.NewInt(0)
		if err := b.UnmarshalText([]byte(e.Value.(string))); err == nil {
			v := ThisInt(b)
			ret = &value{typ: t, val: v}
		} else {
			halt.As(100, "wrong integer")
		}
	case types.BOOLEAN:
		v := e.Value.(bool)
		ret = &value{typ: t, val: v}
	case types.CHAR:
		var v rune
		switch x := e.Value.(type) {
		case int32:
			v = rune(x)
		case int:
			v = rune(x)
		default:
			halt.As(100, "unsupported rune coding")
		}
		ret = &value{typ: t, val: v}
	case types.STRING:
		v := e.Value.(string)
		ret = &value{typ: t, val: v}
	case types.TRILEAN:
		if e.Value == nil {
			ret = &value{typ: t, val: tri.NIL}
		} else if tv := e.Value.(bool); tv {
			ret = &value{typ: t, val: tri.TRUE}
		} else {
			ret = &value{typ: t, val: tri.FALSE}
		}
	case types.REAL:
		r := big.NewRat(0, 1)
		if err := r.UnmarshalText([]byte(e.Value.(string))); err == nil {
			v := ThisRat(r)
			ret = &value{typ: t, val: v}
		} else {
			halt.As(100, "wrong real")
		}
	default:
		halt.As(100, "unknown type ", t, " for ", e)
	}
	return
}
