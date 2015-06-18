package target

import (
	"fmt"
	"github.com/kpmy/ypk/assert"
	"io"
	"leaf/ir"
)

var Ext func(*ir.Module, io.Writer)
var Int func(io.Reader) *ir.Module

func New(mod *ir.Module, tg io.Writer) {
	fmt.Println("MODULE", mod.Name)
	/*for k, v := range mod.ConstDecl {
		fmt.Println("CONST", k, v)
	}
	for k, v := range mod.VarDecl {
		fmt.Println("VAR", k, v)
	}
	fmt.Println("BEGIN")
	for _, v := range mod.BeginSeq {
		fmt.Println(v)
	}
	fmt.Println("CLOSE")
	for _, v := range mod.CloseSeq {
		fmt.Println(v)
	}*/
	fmt.Println("END", mod.Name)
	assert.For(Ext != nil, 20)
	Ext(mod, tg)
}

func Old(sc io.Reader) *ir.Module {
	assert.For(Int != nil, 20)
	return Int(sc)
}
