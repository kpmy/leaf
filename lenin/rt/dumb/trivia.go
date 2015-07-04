package dumb

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
	"leaf/lenin/rt"
	"leaf/lenin/trav"
	"leaf/lss"
	"math/big"
	"reflect"
	"runtime"
	"unicode"
)

var Heap *heap

func bi(x int64) *big.Int {
	return big.NewInt(x)
}

//INC x to n
func inc(ctx rt.Context, s rt.Storage, calc rt.Calc) {
	x := s.Get("x")
	n := s.Get("n")
	zero := calc(types.INTEGER, n, operation.Eq, types.INTEGER, trav.NewInt(0), types.BOOLEAN).(bool)
	if zero {
		n = trav.NewInt(1)
	}
	s.Set("x", calc(types.INTEGER, x, operation.Sum, types.INTEGER, n, types.INTEGER))
}

//DEC x to n
func dec(ctx rt.Context, s rt.Storage, calc rt.Calc) {
	x := s.Get("x")
	n := s.Get("n")
	zero := calc(types.INTEGER, n, operation.Eq, types.INTEGER, trav.NewInt(0), types.BOOLEAN).(bool)
	if zero {
		n = trav.NewInt(1)
	}
	s.Set("x", calc(types.INTEGER, x, operation.Diff, types.INTEGER, n, types.INTEGER))
}

//CAP x to cap
func toUpper(ctx rt.Context, s rt.Storage, calc rt.Calc) {
	x := s.Get("x").(rune)
	s.Set("cap", unicode.ToUpper(x))
}

func length(ctx rt.Context, st rt.Storage, calc rt.Calc) {
	t, x := st.Get("in").(*trav.Any).This()
	switch t {
	case types.STRING:
		s := x.(string)
		n := trav.NewInt(int64(len(s)))
		st.Set("out", n)
	case types.LIST:
		l := x.(*trav.List)
		n := trav.NewInt(int64(l.Len()))
		st.Set("out", n)
	default:
		halt.As(100, t)
	}
}

func odd(ctx rt.Context, st rt.Storage, calc rt.Calc) {
	_x := st.Get("in")
	if i := _x.(*trav.Int); i != nil {
		x := &big.Int{}
		*x = i.Int
		cmp := bi(0).Mod(x, bi(2)).Cmp(bi(0))
		st.Set("out", cmp != 0)
	} else {
		halt.As(100, "not an integer ", reflect.TypeOf(i))
	}
}

func resize(ctx rt.Context, st rt.Storage, calc rt.Calc) {
	l := st.Get("list").(*trav.List)
	n := st.Get("n").(*trav.Int)
	l.Len(int(n.Int64()))
}

func typeof(ctx rt.Context, st rt.Storage, calc rt.Calc) {
	a := st.Get("in").(*trav.Any)
	t, _ := a.This()
	at := trav.Atom(t.String())
	st.Set("res", at)
}

func incl(ctx rt.Context, st rt.Storage, calc rt.Calc) {
	s := st.Get("set").(*trav.Set)
	a := st.Get("x").(*trav.Any)
	s.Incl(a)
}

func excl(ctx rt.Context, st rt.Storage, calc rt.Calc) {
	s := st.Get("set").(*trav.Set)
	a := st.Get("x").(*trav.Any)
	s.Excl(a)
}

func values(ctx rt.Context, st rt.Storage, calc rt.Calc) {
	s := st.Get("x").(*trav.Any)
	out := st.Get("out").(*trav.List)
	t, x := s.This()
	switch t {
	case types.SET:
		set := x.(*trav.Set)
		sl := set.AsList()
		out.Len(len(sl))
		for i, x := range sl {
			out.SetVal(i, x)
		}
	case types.MAP:
		m := x.(*trav.Map)
		vl := m.AsList()
		out.Len(len(vl))
		for i, x := range vl {
			out.SetVal(i, x)
		}
	default:
		halt.As(100, "unsupported type ", t)
	}
}

func keys(ctx rt.Context, st rt.Storage, calc rt.Calc) {
	m := st.Get("x").(*trav.Map)
	out := st.Get("out").(*trav.List)
	kl := m.Keys()
	out.Len(len(kl))
	for i, x := range kl {
		out.SetVal(i, x)
	}
}

func alloc(ctx rt.Context, st rt.Storage, calc rt.Calc) {
	p := st.Get("p").(*trav.Ptr)
	adr := Heap.New()
	link := &heapy{h: Heap, adr: adr}
	runtime.SetFinalizer(link, func(obj *heapy) {
		fmt.Println("finalize ", fmt.Sprintf("%X", obj.adr))
	})
	p.Init(adr, link)
}

func process(ctx rt.Context, st rt.Storage, calc rt.Calc) {
	in := st.Get("to").(*trav.Map)
	m := Map(in)
	fn := ctx.Handler()
	m = fn(rt.Message(m))
	if m != nil {
		halt.As(100)
	}
}

func init() {
	buf := bytes.NewBufferString(rt.StdDef)
	p := lead.ConnectTo(lss.ConnectTo(bufio.NewReader(buf)), func(string) (*ir.Import, error) {
		halt.As(100, "imports not allowed here")
		return nil, errors.New("not allowed")
	})
	rt.StdImp, _ = p.Import()
	rt.StdProc[rt.Qualident{Mod: "STD", Proc: "INC"}] = inc
	rt.StdProc[rt.Qualident{Mod: "STD", Proc: "DEC"}] = dec
	rt.StdProc[rt.Qualident{Mod: "STD", Proc: "CAP"}] = toUpper
	rt.StdProc[rt.Qualident{Mod: "STD", Proc: "LEN"}] = length
	rt.StdProc[rt.Qualident{Mod: "STD", Proc: "ODD"}] = odd
	rt.StdProc[rt.Qualident{Mod: "STD", Proc: "RESIZE"}] = resize
	rt.StdProc[rt.Qualident{Mod: "STD", Proc: "TYPEOF"}] = typeof
	rt.StdProc[rt.Qualident{Mod: "STD", Proc: "INCL"}] = incl
	rt.StdProc[rt.Qualident{Mod: "STD", Proc: "EXCL"}] = excl
	rt.StdProc[rt.Qualident{Mod: "STD", Proc: "VALUES"}] = values
	rt.StdProc[rt.Qualident{Mod: "STD", Proc: "KEYS"}] = keys
	rt.StdProc[rt.Qualident{Mod: "STD", Proc: "NEW"}] = alloc
	rt.StdProc[rt.Qualident{Mod: "STD", Proc: "PROCESS"}] = process
	Heap = newHeap()
}
