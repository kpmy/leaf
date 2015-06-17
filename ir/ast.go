package ir

import (
	"leaf/ir/operation"
	"leaf/ir/types"
)

type Module struct {
	Name      string
	ConstDecl map[string]*Const
	VarDecl   map[string]*Variable
	BeginSeq  []Statement
	CloseSeq  []Statement
}

type Const struct {
	Name string
	Expr Expression
}

type Variable struct {
	Name string
	Type types.Type
}

type Expression interface {
	Self()
}

type EvaluatedExpression interface {
	Expression
	Eval() Expression
}

type Statement interface {
	Do()
}

type AtomExpr struct {
	Value string
}

func (a *AtomExpr) Self() {}

type ConstExpr struct {
	Type  types.Type
	Value interface{}
}

func (c *ConstExpr) Self() {}

type AssignStmt struct {
	Object *Variable
	Expr   Expression
}

func (a *AssignStmt) Do() {}

type NamedConstExpr struct {
	Named *Const
}

func (e *NamedConstExpr) Self() {}

type VariableExpr struct {
	Obj *Variable
}

func (v *VariableExpr) Self() {}

type Monadic struct {
	Op      operation.Operation
	Operand Expression
}

func (m *Monadic) Self() {}

type Dyadic struct {
	Op          operation.Operation
	Left, Right Expression
}

func (d *Dyadic) Self() {}
