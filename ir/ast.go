package ir

import (
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
	Eval()
}

type Statement interface {
	Do()
}

type AtomExpr struct {
	Value string
}

func (a *AtomExpr) Eval() {}

type ConstExpr struct {
	Type  types.Type
	Value interface{}
}

func (c *ConstExpr) Eval() {}

type AssignStmt struct {
	Object *Variable
	Expr   Expression
}

func (a *AssignStmt) Do() {}

type NamedConstExpr struct {
	Named *Const
}

func (e *NamedConstExpr) Eval() {}
