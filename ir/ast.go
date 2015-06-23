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

func (m *Module) Init() {
	m.ConstDecl = make(map[string]*Const)
	m.VarDecl = make(map[string]*Variable)
}

type Const struct {
	Name string
	Expr Expression
}

type Variable struct {
	Name string
	Type types.Type
}

type Selector interface {
	Select()
}

type ChainSelector interface {
	Selector
	Chain() []Selector
}

type SelectVar struct {
	Var *Variable
}

func (s *SelectVar) Select() {}

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
	Sel  Selector
	Expr Expression
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

type SelectExpr struct {
	Base Expression
	Sel  Selector
}

func (s *SelectExpr) Self() {}

type SelectIndex struct {
	Expr Expression
}

func (s *SelectIndex) Select() {}

type IfStmt struct {
	Cond []*ConditionBranch
	Else *ElseBranch
}

func (i *IfStmt) Do() {}

type WhileStmt struct {
	Cond []*ConditionBranch
}

func (i *WhileStmt) Do() {}

type RepeatStmt struct {
	Cond *ConditionBranch
}

func (i *RepeatStmt) Do() {}

type ConditionBranch struct {
	Expr Expression
	Seq  []Statement
}

type ElseBranch struct {
	Seq []Statement
}
