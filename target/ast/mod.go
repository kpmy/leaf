package ast

import (
	"github.com/kpmy/ypk/halt"
	"leaf/ir"
	"leaf/scanner"
	"leaf/target"
)

type modCons struct {
	mod  *ir.Module
	root *tg
}

func (m *modCons) BeginObject(target.Class) {

}

func (m *modCons) Name(string) {

}

func (m *modCons) Value(scanner.Symbol, ...string) {

}

func (m *modCons) EndObject() {

}

func (m *modCons) BeginStatement(sym scanner.Symbol) {
	switch sym {
		case scanner.Becomes:
			cons:=&assignCons{root: m.root, parent: m, stmt: ir.NewAssignStmt()}
			mod.
	default:
		halt.As(100, sym)
	}
}

func (m *modCons) EndStatement() {

}

func (m *modCons) Select(string) {

}

func (m *modCons) BeginExpression() {

}

func (m *modCons) EndExpression() {

}
