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

	And
	Or
	Not

	Eq
	Neq
	Gtr
	Geq
	Lss
	Leq
)

var ops map[scanner.Symbol]Operation

func init() {
	ops = map[scanner.Symbol]Operation{scanner.Plus: Sum,
		scanner.Minus:  Diff,
		scanner.Times:  Prod,
		scanner.Divide: Quot,
		scanner.Div:    Div,
		scanner.Mod:    Mod,
		scanner.And:    And,
		scanner.Or:     Or,
		scanner.Not:    Not,
		scanner.Equal:  Eq,
		scanner.Geq:    Geq,
		scanner.Gtr:    Gtr,
		scanner.Nequal: Neq,
		scanner.Lss:    Lss,
		scanner.Leq:    Leq}
}

func Map(sym scanner.Symbol) (ret Operation) {
	ret = ops[sym]
	assert.For(ret != Undef, 60)
	return
}
