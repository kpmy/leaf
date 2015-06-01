package tool

import (
	"errors"
	"fmt"
	"github.com/kpmy/ypk/assert"
	"github.com/kpmy/ypk/halt"
	"leaf/ebnf"
	"reflect"
)

type Applicator func(ebnf.Expression) bool

type Tokenizer interface {
	Apply(ebnf.Expression, Applicator) error
}

func Filter(g ebnf.Grammar, std Tokenizer) (filter func(ebnf.Expression, Applicator) error) {
	filter = func(_e ebnf.Expression, fn Applicator) (ret error) {
		switch e := _e.(type) {
		case ebnf.Sequence:
			for i := 0; i < len(e); i++ {
				if err := filter(e[i], fn); err != nil {
					ret = errors.New(fmt.Sprint("EBNF expects ", e[i], err))
					break
				}
			}
		case *ebnf.Repetition:
			ret = filter(e.Body, fn)
		case ebnf.Alternative:
			ret = errors.New(fmt.Sprint("EBNF expects one of ", e))
			for i := 0; i < len(e) && ret != nil; i++ {
				ret = filter(e[i], fn)
			}
		case *ebnf.Name:
			p := g[e.String]
			if p != nil {
				ret = filter(p.Expr, fn)
			} else if ret = std.Apply(e, fn); ret == nil {

			} else {
				halt.As(100, e.String)
			}
		case *ebnf.Token:
			if ok := fn(e); !ok {
				ret = errors.New(fmt.Sprint("EBNF expects ", e))
			}
		default:
			halt.As(100, reflect.TypeOf(e))
		}
		return
	}
	return
}

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
