package ast

import (
	"container/list"
	"github.com/kpmy/ypk/assert"
	"leaf/ir"
	"leaf/scanner"
	"leaf/target"
)

type ErrorMarker interface {
	Mark(...interface{})
}

type tg struct {
	em      ErrorMarker
	modName string
	mod     *ir.Module
	stack   *list.List
}

func (t *tg) push(c target.Consumer) {
	t.stack.PushFront(c)
}

func (t *tg) pop() {
	if t.stack.Len() > 0 {
		t.stack.Remove(t.stack.Front())
	}
}

func (t *tg) top() (ret target.Consumer) {
	if t.stack.Len() > 0 {
		ret = t.stack.Front().Value.(target.Consumer)
	}
	return
}

func (t *tg) Open(name string) {
	t.modName = name
	t.mod = ir.NewMod(name)
	t.push(&modCons{mod: t.mod, root: t})
}

func (t *tg) Close(name string) {
	if name != t.modName {
		t.em.Mark("module name does't match")
	}
}

func (t *tg) BeginObject(c target.Class) {
	assert.For(t.top() != nil, 20)
	t.top().BeginObject(c)
}

func (t *tg) EndObject() {
	assert.For(t.top() != nil, 20)
	t.top().EndObject()
}

func (t *tg) Name(name string) {
	assert.For(t.top() != nil, 20)
	t.top().Name(name)
}

func (t *tg) BeginStatement(sym scanner.Symbol) {
	assert.For(t.top() != nil, 20)
	t.top().BeginStatement(sym)
}

func (t *tg) EndStatement() {
	assert.For(t.top() != nil, 20)
	t.top().EndStatement()
}

func (t *tg) Select(name string) {
	assert.For(t.top() != nil, 20)
	t.top().Select(name)
}

func (t *tg) BeginExpression() {
	assert.For(t.top() != nil, 20)
	t.top().BeginExpression()
}

func (t *tg) EndExpression() {
	assert.For(t.top() != nil, 20)
	t.top().EndExpression()
}

func (t *tg) Value(sym scanner.Symbol, val ...string) {
	assert.For(t.top() != nil, 20)
	t.top().Value(sym, val...)
}

func New(e ErrorMarker) target.Target {
	return &tg{em: e, stack: list.New()}
}
