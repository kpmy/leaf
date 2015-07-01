package ir

import (
	"leaf/ir/modifiers"
	"leaf/ir/operation"
	"leaf/ir/types"
)

type Module struct {
	Name      string
	ImportSeq []*Import
	ConstDecl map[string]*Const
	VarDecl   map[string]*Variable
	ProcDecl  map[string]*Procedure
	BeginSeq  []Statement
	CloseSeq  []Statement
}

func (m *Module) Init() {
	m.ConstDecl = make(map[string]*Const)
	m.VarDecl = make(map[string]*Variable)
	m.ProcDecl = make(map[string]*Procedure)
}

type Procedure struct {
	Name      string
	ConstDecl map[string]*Const
	VarDecl   map[string]*Variable
	ProcDecl  map[string]*Procedure
	Infix     []*Variable
	Seq       []Statement
	Pre, Post []Expression
	Modifier  modifiers.Modifier
}

func (p *Procedure) Init() {
	p.ConstDecl = make(map[string]*Const)
	p.VarDecl = make(map[string]*Variable)
	p.ProcDecl = make(map[string]*Procedure)
}

type Const struct {
	Name     string
	Expr     Expression
	Modifier modifiers.Modifier
}

type Variable struct {
	Name     string
	Type     types.Type
	Modifier modifiers.Modifier
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

type SelectMod struct {
	Mod string
}

func (s *SelectMod) Select() {}

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

type WrappedStatement interface {
	Statement
	Fwd() Statement
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

type CallStmt struct {
	Mod  string
	Proc *Procedure
	Par  []*Parameter
}

func (c *CallStmt) Do() {}

type InvokeStmt struct {
	Mod  string
	Proc string
	Par  []*Parameter
}

func (c *InvokeStmt) Do() {}

type Parameter struct {
	Var  *Variable
	Sel  Selector
	Expr Expression
}

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

type Infix struct {
	Mod  string
	Proc *Procedure
	Args []Expression
	Len  int
}

func (i *Infix) Self() {}

type InvokeInfix struct {
	Mod  string
	Proc string
	Args []Expression
	Len  int
}

func (i *InvokeInfix) Self() {}

type TypeTest struct {
	Typ     types.Type
	Operand Expression
}

func (t *TypeTest) Self() {}

type SelectExpr struct {
	Base          Expression
	Before, After Selector
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

type ChooseStmt struct {
	TypeTest bool
	Expr     Expression
	Cond     []*ConditionBranch
	Else     *ElseBranch
}

func (c *ChooseStmt) Do() {}

type SetExpr struct {
	Expr []Expression
}

func (e *SetExpr) Self() {}

type ListExpr struct {
	Expr []Expression
}

func (e *ListExpr) Self() {}

type MapExpr struct {
	Key, Value []Expression
}

func (e *MapExpr) Self() {}
