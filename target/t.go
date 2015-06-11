package target

import (
	"leaf/scanner"
)

type Class int

const (
	Wrong Class = iota
	Constant
	Variable
)

type Target interface {
	Declare(Class)
	BeginBlock(...scanner.Symbol)
	EndBlock()
}
