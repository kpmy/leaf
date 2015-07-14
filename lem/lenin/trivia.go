package lenin

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"github.com/kpmy/ypk/halt"
	"leaf/ir"
	"leaf/ir/operation"
	"leaf/ir/types"
	"leaf/lead"
	_ "leaf/leap"
	"leaf/lem"
	"leaf/lss"
	"math/big"
	"reflect"
	"runtime"
	"unicode"
)

type heap struct {
	data map[int64]*Any
	next int64
}

func (h *heap) New() (ret int64) {
	h.data[h.next] = &Any{}
	ret = h.next
	h.next++
	return
}

type heapy struct {
	adr int64
	h   *heap
}

func (h *heapy) Get() *Any {
	return h.h.data[h.adr]
}

func (h *heapy) Set(x *Any) {
	if lem.Debug {
		fmt.Println("heap touch", fmt.Sprintf("%X", h.adr), x)
	}
	h.h.data[h.adr] = x
	if lem.Debug {
		fmt.Println(h.h)
	}
}

func newHeap() *heap {
	ret := &heap{}
	ret.data = make(map[int64]*Any)
	ret.next = 4096
	return ret
}

func bi(x int64) *big.Int {
	return big.NewInt(x)
}

//INC x to n
func inc(ctx lem.Context, s lem.Storage, calc lem.Calc, par ...lem.VarPar) {
	x := s.Get("x")
	n := s.Get("n")
	zero := calc(types.INTEGER, n, operation.Eq, types.INTEGER, NewInt(0), types.BOOLEAN).(bool)
	if zero {
		n = NewInt(1)
	}
	s.Set("x", calc(types.INTEGER, x, operation.Sum, types.INTEGER, n, types.INTEGER))
}

//DEC x to n
func dec(ctx lem.Context, s lem.Storage, calc lem.Calc, par ...lem.VarPar) {
	x := s.Get("x")
	n := s.Get("n")
	zero := calc(types.INTEGER, n, operation.Eq, types.INTEGER, NewInt(0), types.BOOLEAN).(bool)
	if zero {
		n = NewInt(1)
	}
	s.Set("x", calc(types.INTEGER, x, operation.Diff, types.INTEGER, n, types.INTEGER))
}

//CAP x to cap
func toUpper(ctx lem.Context, s lem.Storage, calc lem.Calc, par ...lem.VarPar) {
	x := s.Get("x").(rune)
	s.Set("cap", unicode.ToUpper(x))
}

func length(ctx lem.Context, st lem.Storage, calc lem.Calc, par ...lem.VarPar) {
	t, x := st.Get("in").(*Any).This()
	switch t {
	case types.STRING:
		s := x.(string)
		n := NewInt(int64(len(s)))
		st.Set("out", n)
	case types.LIST:
		l := x.(*List)
		n := NewInt(int64(l.Len()))
		st.Set("out", n)
	default:
		halt.As(100, t)
	}
}

func odd(ctx lem.Context, st lem.Storage, calc lem.Calc, par ...lem.VarPar) {
	_x := st.Get("in")
	if i := _x.(*Int); i != nil {
		x := &big.Int{}
		*x = i.Int
		cmp := bi(0).Mod(x, bi(2)).Cmp(bi(0))
		st.Set("out", cmp != 0)
	} else {
		halt.As(100, "not an integer ", reflect.TypeOf(i))
	}
}

func resize(ctx lem.Context, st lem.Storage, calc lem.Calc, par ...lem.VarPar) {
	l := st.Get("list").(*List)
	n := st.Get("n").(*Int)
	l.Len(int(n.Int64()))
}

func typeof(ctx lem.Context, st lem.Storage, calc lem.Calc, par ...lem.VarPar) {
	a := st.Get("in").(*Any)
	t, _ := a.This()
	at := Atom(t.String())
	st.Set("res", at)
}

func incl(ctx lem.Context, st lem.Storage, calc lem.Calc, par ...lem.VarPar) {
	s := st.Get("set").(*Set)
	a := st.Get("x").(*Any)
	s.Incl(a)
}

func excl(ctx lem.Context, st lem.Storage, calc lem.Calc, par ...lem.VarPar) {
	s := st.Get("set").(*Set)
	a := st.Get("x").(*Any)
	s.Excl(a)
}

func values(ctx lem.Context, st lem.Storage, calc lem.Calc, par ...lem.VarPar) {
	s := st.Get("x").(*Any)
	out := st.Get("out").(*List)
	t, x := s.This()
	switch t {
	case types.SET:
		set := x.(*Set)
		sl := set.AsList()
		out.Len(len(sl))
		for i, x := range sl {
			out.SetVal(i, x)
		}
	case types.MAP:
		m := x.(*Map)
		vl := m.AsList()
		out.Len(len(vl))
		for i, x := range vl {
			out.SetVal(i, x)
		}
	default:
		halt.As(100, "unsupported type ", t)
	}
}

func keys(ctx lem.Context, st lem.Storage, calc lem.Calc, par ...lem.VarPar) {
	m := st.Get("x").(*Map)
	out := st.Get("out").(*List)
	kl := m.Keys()
	out.Len(len(kl))
	for i, x := range kl {
		out.SetVal(i, x)
	}
}

func (h *heap) prepare(p *Ptr, initial *Any) {
	adr := h.New()
	link := &heapy{h: h, adr: adr}
	runtime.SetFinalizer(link, func(obj *heapy) {
		fmt.Println("finalize ", fmt.Sprintf("%X", obj.adr))
	})
	if initial != nil {
		link.Set(initial)
	}
	p.Init(adr, link)
}

func alloc(ctx lem.Context, st lem.Storage, calc lem.Calc, par ...lem.VarPar) {
	p := st.Get("p").(*Ptr)
	ctx.(*context).heap.prepare(p, nil)
}

func process(ctx lem.Context, st lem.Storage, calc lem.Calc, par ...lem.VarPar) {
	in := st.Get("to").(*Map)
	m := DoMap(in)
	fn := ctx.Handler()
	m = fn(lem.Message(m))
	if m != nil {
		st.Set("from", Unmap(ctx.(*context).heap, m))
	}
}

func trapIf(ctx lem.Context, st lem.Storage, calc lem.Calc, par ...lem.VarPar) {
	cond := st.Get("cond").(bool)
	msg := st.Get("msg").(*Any)
	code := st.Get("code").(*Int)
	if !cond {
		halt.As(100, code, msg)
	}
}

func trap(ctx lem.Context, st lem.Storage, calc lem.Calc, par ...lem.VarPar) {
	msg := st.Get("msg").(*Any)
	code := st.Get("code").(*Int)
	halt.As(100, code, msg)
}

func run(ctx lem.Context, st lem.Storage, calc lem.Calc, par ...lem.VarPar) {
	proc := st.Get("proc").(*Proc)
	p := proc.This()
	if p != nil {
		ctx.Queue(p, par...)
	}
}

func init() {
	buf := bytes.NewBufferString(lem.StdDef)
	p := lead.ConnectTo(lss.ConnectTo(bufio.NewReader(buf)), func(string) (*ir.Import, error) {
		halt.As(100, "imports not allowed here")
		return nil, errors.New("not allowed")
	})
	lem.StdImp, _ = p.Import()
	lem.StdProc[lem.Qualident{Mod: "STD", Proc: "INC"}] = inc
	lem.StdProc[lem.Qualident{Mod: "STD", Proc: "DEC"}] = dec
	lem.StdProc[lem.Qualident{Mod: "STD", Proc: "CAP"}] = toUpper
	lem.StdProc[lem.Qualident{Mod: "STD", Proc: "LEN"}] = length
	lem.StdProc[lem.Qualident{Mod: "STD", Proc: "ODD"}] = odd
	lem.StdProc[lem.Qualident{Mod: "STD", Proc: "RESIZE"}] = resize
	lem.StdProc[lem.Qualident{Mod: "STD", Proc: "TYPEOF"}] = typeof
	lem.StdProc[lem.Qualident{Mod: "STD", Proc: "INCL"}] = incl
	lem.StdProc[lem.Qualident{Mod: "STD", Proc: "EXCL"}] = excl
	lem.StdProc[lem.Qualident{Mod: "STD", Proc: "VALUES"}] = values
	lem.StdProc[lem.Qualident{Mod: "STD", Proc: "KEYS"}] = keys
	lem.StdProc[lem.Qualident{Mod: "STD", Proc: "NEW"}] = alloc
	lem.StdProc[lem.Qualident{Mod: "STD", Proc: "PROCESS"}] = process
	lem.StdProc[lem.Qualident{Mod: "STD", Proc: "ASSERT"}] = trapIf
	lem.StdProc[lem.Qualident{Mod: "STD", Proc: "HALT"}] = trap
	lem.StdProc[lem.Qualident{Mod: "STD", Proc: "RUN"}] = run
	lem.Special[lem.Qualident{Mod: "STD", Proc: "RUN"}] = lem.Prop{Variadic: true}
}
