package main

import (
	"fmt"
	"github.com/kpmy/ypk/halt"
	"leaf/ebnf"
	"log"
	"os"
	"reflect"
)

var name string = "cp.ebnf"

func main() {
	if f, err := os.Open(name); err == nil {
		defer f.Close()
		passed := make(map[string]interface{})
		var dump func(interface{})
		depth := 0
		if g, err := ebnf.Parse(name, f); err == nil {
			dump = func(_x interface{}) {
				depth++
				switch x := _x.(type) {
				case ebnf.Grammar:
					fmt.Println("grammar")
					for _, v := range x {
						dump(v)
					}
				case *ebnf.Production:
					fmt.Print(x.Name.String, " = ")
					dump(x.Expr)
					fmt.Println()
				case ebnf.Sequence:
					for _, v := range x {
						dump(v)
					}
				case ebnf.Alternative:
					for i, v := range x {
						if i > 0 {
							fmt.Print("|")
						}
						dump(v)
					}
				case *ebnf.Option:
					fmt.Print("[")
					dump(x.Body)
					fmt.Print("]")
				case *ebnf.Repetition:
					fmt.Print("{")
					dump(x.Body)
					fmt.Print("}")
				case *ebnf.Group:
					fmt.Print("(")
					dump(x.Body)
					fmt.Print(")")
				case *ebnf.Token:
					fmt.Print(`'`, x.String, `'`)
				case *ebnf.Name:
					if passed[x.String] == nil {
						p := g[x.String]
						if p != nil && depth < 1 {
							passed[x.String] = x
							dump(p)
						} else {
							fmt.Print(" ", x.String, " ")
						}
					} else {
						fmt.Print(" ", x.String, " ")
					}
				default:
					halt.As(100, reflect.TypeOf(x))
				}
			}
			dump(g)
		} else {
			log.Fatal(err)
		}
	} else {
		log.Fatal(err)
	}
}
