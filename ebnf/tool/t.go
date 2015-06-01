package tool

import (
	"fmt"
	"github.com/kpmy/ypk/halt"
	"leaf/ebnf"
	"reflect"
)

type Item struct {
	this ebnf.Expression
	up   []*Item
	down []*Item
}

func (i *Item) String() (ret string) {
	/*	if i.parent != nil && i.parent.this != nil {
			ret = fmt.Sprint("(", reflect.TypeOf(i.parent.this), ")")
		}
	*/
	switch e := i.this.(type) {
	case *ebnf.Name:
		ret = fmt.Sprint(e.String)
	case *ebnf.Token:
		ret = fmt.Sprint(`"`, e.String, `"`)
	default:
		ret = fmt.Sprint(ret, reflect.TypeOf(i.this))
	}
	/*	for _, v := range i.children {
			ret = fmt.Sprint(ret, v)
		}
	*/
	return
}

func reorder(i *Item) *Item {
	fmt.Println("REORDER")
	depth := 0
	dm := make(map[int][]*Item)
	done := make([]*Item, 0)

	isDone := func(i *Item) bool {
		for _, x := range done {
			if i == x {
				return true
			}
		}
		return false
	}

	var dump func(i *Item)

	dump = func(i *Item) {
		if !isDone(i) {
			dm[depth] = append(dm[depth], i)
		}
		depth++
		fmt.Println(len(i.up), len(i.down))
		for _, x := range i.down {
			dump(x)
		}
		depth--
	}

	dump(i)
	for k, v := range dm {
		fmt.Println(k, v)
	}
	fmt.Println("REORDERED")
	return nil
}

func join(up, down *Item) {
	exists := func(list []*Item, i *Item) bool {
		for _, x := range list {
			if i == x {
				return true
			}
		}
		return false
	}

	if up != nil && down != nil {
		if !exists(up.down, down) {
			up.down = append(up.down, down)
		}
		if !exists(down.up, up) {
			down.up = append(down.up, up)
		}
	}
}

func Transform(g ebnf.Grammar) *Item {
	passed := make(map[string]*Item)
	var dump func(*Item, interface{}) *Item

	dump = func(root *Item, _x interface{}) (ret *Item) {
		switch x := _x.(type) {
		case ebnf.Grammar:
			ret = &Item{}
			join(root, ret)
			for _, v := range x {
				dump(ret, v)
			}
		case *ebnf.Production:
			dump(root, x.Expr)
		case ebnf.Sequence:
			ret = &Item{this: x}
			join(root, ret)
			for _, v := range x {
				dump(ret, v)
			}
		case ebnf.Alternative:
			ret = &Item{this: x}
			join(root, ret)
			for _, v := range x {
				dump(ret, v)
			}
		case *ebnf.Option:
			ret = &Item{this: x}
			join(root, ret)
			dump(ret, x.Body)
		case *ebnf.Repetition:
			ret = &Item{this: x}
			join(root, ret)
			dump(ret, x.Body)
		case *ebnf.Group:
			ret = &Item{this: x}
			join(root, ret)
			dump(ret, x.Body)
		case *ebnf.Token:
			ret = &Item{this: x}
			join(root, ret)
		case *ebnf.Name:
			if passed[x.String] == nil {
				p := g[x.String]
				if p != nil {
					passed[x.String] = dump(root, p)
				} else {
					ret = &Item{this: x}
					join(root, ret)
				}
			} else {
				join(root, passed[x.String])
			}
		default:
			halt.As(100, reflect.TypeOf(x))
		}
		return
	}
	ret := dump(nil, g)
	return reorder(ret)
}
