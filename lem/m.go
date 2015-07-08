package lem

import (
	"github.com/kpmy/ypk/assert"
	"leaf/lenin/rt"
	"leaf/lenin/trav"
)

type Object interface {
	Value() *trav.Any
}

type Machine interface {
	Chan() chan rt.Message
	Stop()
}

var Rt func() Machine

func Run() Machine {
	assert.For(Rt != nil, 20)
	return Rt()
}
