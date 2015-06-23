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
	Pow

	Im
	Pcmp
	Ncmp

	And
	Or
	Not

	Eq
	Neq
	Gtr
	Geq
	Lss
	Leq
	//leave this last
	None
)

var ops map[scanner.Symbol]Operation
var OpMap map[string]Operation

func (o Operation) String() string {
	switch o {
	case Neg:
		return "--"
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
	case Pow:
		return "^"
	case Im:
		return "!"
	case Ncmp:
		return "-!"
	case Pcmp:
		return "+!"
	case None:
		return "nop"
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
		scanner.Leq:    Leq,
		scanner.Arrow:  Pow,
		scanner.Ncmp:   Ncmp,
		scanner.Pcmp:   Pcmp}

	OpMap = make(map[string]Operation)
	for i := int(Undef); i < int(None); i++ {
		OpMap[Operation(i).String()] = Operation(i)
	}
}

func Map(sym scanner.Symbol) (ret Operation) {
	ret = ops[sym]
	assert.For(ret != Undef, 60)
	return
}
