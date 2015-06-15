package operation

import (
	"github.com/kpmy/ypk/assert"
	"leaf/scanner"
)

type Operation int

const (
	Undef Operation = iota
	Neg
	Sum
	Diff
	Prod
	Quot
	Div
	Mod
)

var ops map[scanner.Symbol]Operation

func init() {
	ops = map[scanner.Symbol]Operation{scanner.Plus: Sum,
		scanner.Minus:  Diff,
		scanner.Times:  Prod,
		scanner.Divide: Quot,
		scanner.Div:    Div,
		scanner.Mod:    Mod}
}

func Map(sym scanner.Symbol) (ret Operation) {
	ret = ops[sym]
	assert.For(ret != Undef, 60)
	return
}
