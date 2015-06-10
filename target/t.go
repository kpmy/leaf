package target

import (
	"leaf/scanner"
)

type Class int

const (
	Wrong Class = iota
	Variable
)

type Target interface {
	Open(string)
	BeginObject(Class)
	Name(string)
	EndObject()
	Close(string)
	BeginStatement(scanner.Symbol)
	EndStatement()
	Select(string)
	BeginExpression()
	EndExpression()
}
