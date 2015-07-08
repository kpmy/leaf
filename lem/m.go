package lem

import (
	"github.com/kpmy/ypk/assert"
	"leaf/lenin/rt"
)

type Object interface {
	Convert()
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
