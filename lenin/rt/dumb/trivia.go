package dumb

import (
	"bufio"
	"bytes"
	"errors"
	"github.com/kpmy/ypk/halt"
	"leaf/ir"
	"leaf/ir/operation"
	"leaf/ir/types"
	"leaf/lead"
	_ "leaf/leap"
	"leaf/lenin/rt"
	"leaf/lenin/trav"
	"leaf/lss"
	"unicode"
)

//INC x to n
func inc(s rt.Storage, calc rt.Calc) {
	x := s.Get("x")
	n := s.Get("n")
	zero := calc(types.INTEGER, n, operation.Eq, types.INTEGER, trav.NewInt(0), types.BOOLEAN).(bool)
	if zero {
		n = trav.NewInt(1)
	}
	s.Set("x", calc(types.INTEGER, x, operation.Sum, types.INTEGER, n, types.INTEGER))
}

//DEC x to n
func dec(s rt.Storage, calc rt.Calc) {
	x := s.Get("x")
	n := s.Get("n")
	zero := calc(types.INTEGER, n, operation.Eq, types.INTEGER, trav.NewInt(0), types.BOOLEAN).(bool)
	if zero {
		n = trav.NewInt(1)
	}
	s.Set("x", calc(types.INTEGER, x, operation.Diff, types.INTEGER, n, types.INTEGER))
}

//CAP x to cap
func toUpper(s rt.Storage, calc rt.Calc) {
	x := s.Get("x").(rune)
	s.Set("cap", unicode.ToUpper(x))
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
}
