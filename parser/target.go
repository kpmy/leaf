package parser

import (
	"github.com/kpmy/ypk/halt"
	"leaf/ir"
)

type target struct {
	root *ir.Module
}

func (t *target) init(mod string) {
	t.root = &ir.Module{Name: mod}
	t.root.ConstDecl = make(map[string]*ir.Const)
	t.root.VarDecl = make(map[string]*ir.Variable)
}

type scopeLevel struct {
	varScope   map[string]*ir.Variable
	constScope map[string]*ir.Const
}

type exprBuilder struct {
	scope scopeLevel
	stack []ir.Expression
}

func (e *exprBuilder) Eval() {}

func (e *exprBuilder) factor(expr ir.Expression) {
	e.stack = append(e.stack, expr)
}

func (e *exprBuilder) as(id string) ir.Expression {
	if c := e.scope.constScope[id]; c != nil {
		return &ir.NamedConstExpr{Named: c}
	} else if v := e.scope.varScope[id]; v != nil {
		halt.As(100)
	}
	panic(0)
}

type blockBuilder struct {
	scope scopeLevel
	seq   []ir.Statement
}

func (b *blockBuilder) obj(id string) *ir.Variable {
	return b.scope.varScope[id]
}

func (b *blockBuilder) put(s ir.Statement) {
	b.seq = append(b.seq, s)
}
