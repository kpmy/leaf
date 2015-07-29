package lem

import (
	"github.com/kpmy/ypk/assert"
	"leaf/ir"
)

type Context interface {
	Handler() func(Message) Message
	Queue(interface{}, ...VarPar)
}

type Storage interface {
	List() []*ir.Variable
	Set(string, interface{})
	Get(string) interface{}
}

type Loader func(string) (*ir.Module, error)

type Message map[interface{}]interface{}

type Object interface {
	Value() interface{}
}

type Machine interface {
	Input() chan Message
	Chan() chan Message
	Do(m *ir.Module, ld Loader)
	Stop()
}

var Rt func() Machine

func Start() Machine {
	assert.For(Rt != nil, 20)
	return Rt()
}

var Debug = false
