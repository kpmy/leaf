package tool

import (
	"github.com/kpmy/ypk/assert"
	"github.com/kpmy/ypk/halt"
	"leaf/ebnf"
	"reflect"
)

func Top(g ebnf.Grammar) (ret *ebnf.Production) {
	passed := make(map[string]interface{})
	invalid := make(map[string]interface{})

	var dump func(interface{})
	depth := 0
	dump = func(_x interface{}) {
		depth++
		switch x := _x.(type) {
		case ebnf.Grammar:
			for _, v := range x {
				dump(v)
			}
		case *ebnf.Production:
			dump(x.Expr)
		case ebnf.Sequence:
			for _, v := range x {
				dump(v)
			}
		case ebnf.Alternative:
			for _, v := range x {
				dump(v)
			}
		case *ebnf.Option:
			dump(x.Body)
		case *ebnf.Repetition:
			dump(x.Body)
		case *ebnf.Group:
			dump(x.Body)
		case *ebnf.Token:
		case *ebnf.Name:
			if passed[x.String] == nil {
				p := g[x.String]
				passed[x.String] = x
				if p != nil && depth < 1 {
					dump(p)
				} else {
					if p == nil {
						invalid[x.String] = x
					}
				}
			} else {
			}
		default:
			halt.As(100, reflect.TypeOf(x))
		}
	}
	dump(g)
	for k, v := range g {
		if passed[k] == nil {
			ret = v
		}
	}
	assert.For(ret != nil, 20)
	return
}
