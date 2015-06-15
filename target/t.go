package target

import (
	"fmt"
)

import (
	"leaf/ir"
)

func Do(mod *ir.Module) {
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
}
