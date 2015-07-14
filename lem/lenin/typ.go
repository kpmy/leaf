package lenin

import (
	"fmt"
	"github.com/kpmy/trigo"
	"github.com/kpmy/ypk/assert"
	"leaf/ir"
	"leaf/ir/operation"
	"leaf/ir/types"
	"math/big"
)

type Proc struct {
	p *ir.Procedure
}

func (p *Proc) This() *ir.Procedure {
	return p.p
}

func (p *Proc) String() string {
	if p.p != nil {
		return fmt.Sprint("@", p.p.Name)
	} else {
		return fmt.Sprint("@", "undef")
	}
}

func NewProc(p *ir.Procedure) *Proc {
	return &Proc{p: p}
}

func ThisProc(p *Proc) *Proc {
	ret := &Proc{}
	ret.p = p.p
	return ret
}

type Extractor interface {
	Get() *Any
	Set(*Any)
}

type Ptr struct {
	adr  int64
	link Extractor
}

func ThisPtr(p *Ptr) (ret *Ptr) {
	ret = &Ptr{}
	ret.adr = p.adr
	ret.link = p.link
	return
}

func (p *Ptr) Init(x int64, link Extractor) {
	p.adr = x
	p.link = link
}

func (p *Ptr) Get() *Any {
	if p.adr != 0 {
		return p.link.Get()
	} else {
		return nil
	}
}

func (p *Ptr) String() string {
	if p.adr != 0 {
		return fmt.Sprint("$", fmt.Sprintf("%x", p.adr), ": ", p.link.Get())
	} else {
		return "nil"
	}
}

type Map struct {
	k []*Any
	v []*Any
}

func (m *Map) Keys() []*Any {
	return m.k
}

func (m *Map) AsList() (ret []*Any) {
	return m.v
}

func (m *Map) String() (ret string) {
	for i, x := range m.k {
		if i > 0 {
			ret = fmt.Sprint(ret, ", ")
		}
		ret = fmt.Sprint(ret, x, ":", m.v[i])
	}
	return fmt.Sprint("<", ret, ">")
}

func ThisMap(m *Map) (ret *Map) {
	ret = &Map{}
	for _, k := range m.k {
		n := &Any{typ: k.typ, x: k.x}
		ret.k = append(ret.k, n)
	}
	for _, v := range m.v {
		n := &Any{typ: v.typ, x: v.x}
		ret.v = append(ret.v, n)
	}
	return
}

func NewMap(_k, _v []*value) (ret *Map) {
	ret = &Map{}
	for _, k := range _k {
		ret.k = append(ret.k, ThisAny(k))
	}
	for _, v := range _v {
		ret.v = append(ret.v, ThisAny(v))
	}
	return
}

func (m *Map) In(a *Any) (idx int) {
	assert.For(a != nil, 20)
	idx = -1
	for i, x := range m.k {
		if x.Equal(a) {
			idx = i
			break
		}
	}
	return
}

func (m *Map) Set(i *Any, a *Any) {
	if x := m.In(i); x >= 0 {
		m.v[x] = &Any{typ: a.typ, x: a.x}
	} else {
		m.k = append(m.k, &Any{typ: i.typ, x: i.x})
		m.v = append(m.v, &Any{typ: a.typ, x: a.x})
	}
}

func (m *Map) Get(i *Any) (ret *Any) {
	ret = &Any{}
	if x := m.In(i); x >= 0 {
		n := &Any{typ: m.v[x].typ, x: m.v[x].x}
		ret = n
	}
	return
}

type Set struct {
	inv bool
	x   []*Any
}

func (s *Set) String() (ret string) {
	for i, x := range s.x {
		if i > 0 {
			ret = fmt.Sprint(ret, ", ")
		}
		ret = fmt.Sprint(ret, x)
	}
	if s.inv {
		return fmt.Sprint("}", ret, "{")
	} else {
		return fmt.Sprint("{", ret, "}")
	}
}

func (s *Set) In(a *Any) (idx int) {
	assert.For(a != nil, 20)
	idx = -1
	for i, x := range s.x {
		if x.Equal(a) {
			idx = i
			break
		}
	}
	return
}

func (s *Set) Incl(a *Any) {
	assert.For(a.x != nil, 20)
	if s.In(a) < 0 {
		s.x = append(s.x, a)
	}
}

func (s *Set) Excl(a *Any) {
	assert.For(a.x != nil, 20)
	var tmp []*Any
	if i := s.In(a); i >= 0 {
		for idx, x := range s.x {
			if idx != i {
				tmp = append(tmp, x)
			}
		}
		s.x = tmp
	}
}

func (s *Set) Sum(x *Set) {
	assert.For(x != nil, 20)
	for _, v := range x.x {
		s.Incl(v)
	}
}

func (s *Set) Diff(x *Set) {
	assert.For(x != nil, 20)
	for _, v := range x.x {
		s.Excl(v)
	}
}

func (s *Set) Prod(x *Set) {
	assert.For(x != nil, 20)
	for _, v := range s.x {
		if x.In(v) < 0 {
			s.Excl(v)
		}
	}
}

func (s *Set) Quot(x *Set) {
	assert.For(x != nil, 20)
	tmp := &Set{}
	tmp.Sum(s)
	tmp.Prod(x)
	for _, v := range x.x {
		if tmp.In(v) < 0 {
			s.Incl(v)
		}
	}
	for _, v := range s.x {
		if tmp.In(v) >= 0 {
			s.Excl(v)
		}
	}
}

func (s *Set) AsList() (ret []*Any) {
	for _, x := range s.x {
		ret = append(ret, x)
	}
	return
}

func (s *Set) IsEmpty() bool {
	return len(s.x) == 0
}

func NewSet(v ...*value) (s *Set) {
	s = &Set{}
	for _, x := range v {
		s.Incl(ThisAny(x))
	}
	return
}

func ThisSet(s *Set) (ret *Set) {
	ret = &Set{}
	ret.inv = s.inv
	for _, i := range s.x {
		n := &Any{typ: i.typ, x: i.x}
		ret.x = append(ret.x, n)
	}
	return
}

type List struct {
	x []*Any
}

func (l *List) Len(n ...int) int {
	if len(n) == 1 {
		ln := n[0]
		if ln == 0 {
			l.x = nil
		} else if len(l.x) > ln {
			var tmp []*Any
			for _, x := range l.x {
				tmp = append(tmp, x)
			}
			l.x = tmp
		} else if len(l.x) < ln {
			for i := len(l.x); i < ln; i++ {
				l.x = append(l.x, &Any{})
			}
		}
	}
	return len(l.x)
}

func (l *List) Set(i int, x *value) {
	n := &Any{}
	if x.typ == types.ANY {
		t, d := x.toAny().This()
		n.typ, n.x = t, d
	} else {
		n.typ = x.typ
		n.x = x.val
	}
	l.x[i] = n
}

func (l *List) SetVal(i int, x *Any) {
	l.Set(i, &value{typ: types.ANY, val: x})
}

func (l *List) Get(i int) *Any {
	return l.x[i]
}

func (l *List) String() (ret string) {
	for i, x := range l.x {
		if i > 0 {
			ret = fmt.Sprint(ret, ", ")
		}
		ret = fmt.Sprint(ret, x)
	}
	return fmt.Sprint("[", ret, "]")
}

func ThisList(l *List) (ret *List) {
	ret = &List{}
	for _, i := range l.x {
		n := &Any{typ: i.typ, x: i.x}
		ret.x = append(ret.x, n)
	}
	return
}

func NewList(v ...*value) (s *List) {
	s = &List{}
	for _, x := range v {
		s.x = append(s.x, ThisAny(x))
	}
	return
}

type Any struct {
	typ types.Type
	x   interface{}
}

func (a *Any) This() (types.Type, interface{}) {
	return a.typ, a.x
}

func (a *Any) String() string {
	return fmt.Sprint("^", a.x)
}

func (a *Any) Equal(b *Any) (ok bool) {
	ok = false
	if a.x != nil && b.x != nil {
		if a.typ == b.typ {
			v := calcDyadic(&value{typ: a.typ, val: a.x}, operation.Eq, &value{typ: b.typ, val: b.x})
			ok = v.toBool()
		}
	}
	return
}

func ThisAny(v *value) (ret *Any) {
	assert.For(v != nil, 20)
	if a, ok := v.val.(*Any); ok {
		ret = &Any{typ: a.typ, x: a.x}
	} else {
		ret = &Any{typ: v.typ, x: v.val}
	}
	return
}

func NewAny(typ types.Type, val interface{}) *Any {
	_, ok := val.(*Any)
	assert.For(!ok, 20)
	return &Any{typ: typ, x: val}
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
	case propose == types.ANY && expect == types.PROC:
		ret = true
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
	case v.typ == types.ANY && target == types.PROC:
		a := v.toAny()
		assert.For(a.x == nil, 20)
		ret = &value{typ: types.PROC, val: &Proc{}}
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
	case v.typ == types.PTR && target == types.PTR:
		ret = &value{typ: types.PTR, val: ThisPtr(v.val.(*Ptr))} //pointers have only value, not an identity
	case v.typ == target:
		ret = v
	}
	assert.For(ret != nil, 60, v.typ, target, v.val)
	return
}
