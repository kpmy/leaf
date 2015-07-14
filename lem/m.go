package lem

import (
	"github.com/kpmy/ypk/assert"
	"leaf/ir"
	"leaf/ir/operation"
	"leaf/ir/types"
)

type Qualident struct {
	Mod  string
	Proc string
}

type Context interface {
	Handler() func(Message) Message
	Queue(interface{}, ...VarPar)
}

type Storage interface {
	List() []*ir.Variable
	Set(string, interface{})
	Get(string) interface{}
}

type Prop struct {
	Variadic bool
}

type VarPar struct {
	Name string
	Val  interface{}
	Sel  ir.Selector
}

type Calc func(types.Type, interface{}, operation.Operation, types.Type, interface{}, types.Type) interface{}
type Proc func(Context, Storage, Calc, ...VarPar)

var StdImp *ir.Import
var StdProc map[Qualident]Proc
var Special map[Qualident]Prop

type Loader func(string) (*ir.Module, error)

type Message map[interface{}]interface{}

type Object interface {
	Value() interface{}
}

type Machine interface {
	Chan() chan Message
	Do(m *ir.Module, ld Loader)
	Stop()
}

var Rt func() Machine

func Start() Machine {
	assert.For(Rt != nil, 20)
	return Rt()
}

func init() {
	StdImp = &ir.Import{}
	StdImp.Init()
	StdProc = make(map[Qualident]Proc)
	Special = make(map[Qualident]Prop)
}
