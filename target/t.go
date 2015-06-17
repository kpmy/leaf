package target

import (
	"fmt"
	"github.com/kpmy/ypk/assert"
	"io"
	"leaf/ir"
)

var Code func(*ir.Module, io.Writer)

func Do(mod *ir.Module, tg io.Writer) {
	fmt.Println("MODULE", mod.Name)
	for k, v := range mod.ConstDecl {
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
	}
	fmt.Println("END", mod.Name)
	assert.For(Code != nil, 20)
	Code(mod, tg)
}
