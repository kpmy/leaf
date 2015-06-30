package trav

import (
	"fmt"
	"github.com/kpmy/trigo"
	"github.com/kpmy/ypk/assert"
	"github.com/kpmy/ypk/halt"
	"leaf/ir/types"
	"math/big"
)

type Any struct {
	typ types.Type
	x   interface{}
}

func (a *Any) This() (types.Type, interface{}) {
	return a.typ, a.x
}

func (a *Any) String() string {
	return fmt.Sprint(a.x)
}

func ThisAny(v *value) *Any {
	assert.For(v != nil, 20)
	if _, ok := v.val.(*Any); ok {
		halt.As(100)
	}
	return &Any{typ: v.typ, x: v.val}
}

type Atom string

type Int struct {
	big.Int
}

func NewInt(x int64) (ret *Int) {
	ret = &Int{}
	ret.Int = *big.NewInt(x)
	return
}

func ThisInt(x *big.Int) (ret *Int) {
	ret = &Int{}
	ret.Int = *x
	return
}

func (i *Int) String() string {
	x, _ := i.Int.MarshalText()
	return string(x)
}

type Rat struct {
	big.Rat
}

func NewRat(x float64) (ret *Rat) {
	ret = &Rat{}
	ret.Rat = *big.NewRat(0, 1)
	return
}

func ThisRat(x *big.Rat) (ret *Rat) {
	ret = &Rat{}
	ret.Rat = *x
	return
}

type Cmp struct {
	re, im *big.Rat
}

func (c *Cmp) String() (ret string) {
	null := big.NewRat(0, 1)
	if c.re.Cmp(null) != 0 {
		ret = fmt.Sprint(c.re)
	}
	if eq := c.im.Cmp(null); eq > 0 {
		ret = fmt.Sprint(ret, "+i", c.im.Abs(c.im))
	} else if eq < 0 {
		ret = fmt.Sprint(ret, "-i", c.im.Abs(c.im))
	} else if ret == "" {
		ret = "0"
	}
	return
}

func NewCmp(re, im float64) (ret *Cmp) {
	ret = &Cmp{}
	ret.re = big.NewRat(0, 1).SetFloat64(re)
	ret.im = big.NewRat(0, 1).SetFloat64(im)
	return
}

func ThisCmp(c *Cmp) (ret *Cmp) {
	ret = &Cmp{}
	*ret = *c
	return
}

func compTypes(propose, expect types.Type) (ret bool) {
	switch {
	case propose == types.INTEGER && expect == types.REAL:
		ret = true
	case propose == types.BOOLEAN && expect == types.TRILEAN:
		ret = true
	case expect == types.ANY:
		ret = true
	case propose == expect:
		ret = true
	}
	return
}

func conv(v *value, target types.Type) (ret *value) {
	switch {
	case v.typ == types.INTEGER && target == types.REAL:
		i := v.toInt()
		x := big.NewRat(0, 1)
		ret = &value{typ: target, val: ThisRat(x.SetInt(i))}
	case v.typ == types.BOOLEAN && target == types.TRILEAN:
		b := v.toBool()
		x := tri.This(b)
		ret = &value{typ: target, val: x}
	case target == types.ANY && v.typ != types.ANY:
		x := ThisAny(v)
		ret = &value{typ: target, val: x}
	case target == types.ANY && v.typ == types.ANY:
		x := v.toAny()
		ret = &value{typ: target, val: x}
	case v.typ == target:
		ret = v
	}
	assert.For(ret != nil, 60, v.typ, target, v.val)
	return
}
