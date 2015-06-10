package ast

import (
	"container/list"
	"github.com/kpmy/ypk/assert"
	"github.com/kpmy/ypk/halt"
	"leaf/ir"
	"leaf/scanner"
	"leaf/target"
)

type ErrorMarker interface {
	Mark(...interface{})
}

type scope struct {
	sc   *ir.Scope
	this ir.Object
}

type tg struct {
	em      ErrorMarker
	modName string
	mod     *ir.Module
	stack   *list.List
}

func (t *tg) push(s *ir.Scope) {
	t.stack.PushFront(&scope{sc: s})
}

func (t *tg) pop() {
	if t.stack.Len() > 0 {
		t.stack.Remove(t.stack.Front())
	}
}

func (t *tg) top() (ret *scope) {
	if t.stack.Len() > 0 {
		ret = t.stack.Front().Value.(*scope)
	}
	return
}

func (t *tg) Open(name string) {
	t.modName = name
	t.mod = ir.NewMod(name)
	t.push(t.mod.Top)
}

func (t *tg) Close(name string) {
	if name != t.modName {
		t.em.Mark("module name does't match")
	}
}

func (t *tg) BeginObject(c target.Class) {
	assert.For(t.top().this == nil, 20)
	switch c {
	case target.Variable:
		t.top().this = ir.NewVar()
		t.top().sc.Add(t.top().this)
	default:
		halt.As(100, "неизвестный класс ", c)
	}
}

func (t *tg) EndObject() {
	assert.For(t.top().this != nil, 20)
	t.top().this = nil
}

func (t *tg) Name(name string) {
	assert.For(t.top().this != nil, 20)
	t.top().this.Name(name)
}

func (t *tg) BeginStatement(scanner.Symbol) {

}

func (t *tg) EndStatement() {

}

func (t *tg) Select(name string) {

}

func (t *tg) BeginExpression() {

}

func (t *tg) EndExpression() {

}

func New(e ErrorMarker) target.Target {
	return &tg{em: e, stack: list.New()}
}
