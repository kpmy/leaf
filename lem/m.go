package lem

import (
	"github.com/kpmy/ypk/assert"
	"io"
	"leaf/ir"
	"leaf/lenin/rt"
)

type Machine interface {
	Chan() chan rt.Message
}

var Ext func(*ir.Module, io.Writer)
var Int func(io.Reader) *ir.Module
var Rt func() Machine

func New(mod *ir.Module, tg io.Writer) {
	assert.For(Ext != nil, 20)
	Ext(mod, tg)
}

func Old(sc io.Reader) *ir.Module {
	assert.For(Int != nil, 20)
	return Int(sc)
}

func Run() Machine {
	assert.For(Rt != nil, 20)
	return Rt()
}
