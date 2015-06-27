package ir

import (
	"leaf/ir/modifiers"
	"leaf/ir/types"
)

type Import struct {
	Name      string
	ConstDecl map[string]ImportConst
	VarDecl   map[string]ImportVariable
	ProcDecl  map[string]ImportProcedure
}

func (i *Import) Init() {
	i.ConstDecl = make(map[string]ImportConst)
	i.VarDecl = make(map[string]ImportVariable)
	i.ProcDecl = make(map[string]ImportProcedure)
}

type ImportProcedure interface {
	Name() string
	VarDecl() map[string]ImportVariable
	Infix() []ImportVariable
	Pre() []Expression
	Post() []Expression
	This() *Procedure
}

type ImportConst interface {
	Name() string
	Expr() Expression
	This() *Const
}

type ImportVariable interface {
	Name() string
	Type() types.Type
	Modifier() modifiers.Modifier
	This() *Variable
}
