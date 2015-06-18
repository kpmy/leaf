package operation

import (
	"github.com/kpmy/ypk/assert"
	"leaf/scanner"
	"strconv"
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

func (o Operation) String() string {
	switch o {
	case Neg:
		return "-"
	case Sum:
		return "+"
	case Diff:
		return "-"
	case Prod:
		return "*"
	case Quot:
		return "/"
	case Div:
		return "//"
	case Mod:
		return "%"
	case And:
		return "&"
	case Or:
		return "|"
	case Not:
		return "~"
	case Eq:
		return "="
	case Neq:
		return "#"
	case Gtr:
		return ">"
	case Geq:
		return ">="
	case Lss:
		return "<"
	case Leq:
		return "<="
	default:
		return strconv.Itoa(int(o))
	}
}
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
