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
	done := make([]*Item, 0)

	isDone := func(i *Item) bool {
		for _, x := range done {
			if i == x {
				return true
			}
		}
		return false
	}

	var dump func(root *Item, i *Item)

	dump = func(root *Item, i *Item) {
		if !isDone(i) {
			done = append(done, i)
		}
		if len(i.up) == 0 {
			root.down = append(root.down, i)
			for _, v := range i.down {
				dump(root, v)
			}
		}
	}

	top := &Item{}
	dump(top, i)
	for _, v := range top.down {
		fmt.Println(v)
	}
	fmt.Println("REORDERED")
	return top
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
		if !exists(down.up, up) && up.this != nil {
			down.up = append(down.up, up)
		}
	}
}

func Transform(g ebnf.Grammar) *Item {
	passed := make(map[string]*Item)
	cache := make(map[string]ebnf.Expression)
	var dump func(*Item, *Item, interface{})

	dump = func(root *Item, n *Item, _x interface{}) {
		switch x := _x.(type) {
		case ebnf.Grammar:
			for _, v := range x {
				dump(n, &Item{}, v)
			}
		case *ebnf.Production:
			dump(root, n, x.Expr)
		case ebnf.Sequence:
			n.this = x
			join(root, n)
			for _, v := range x {
				dump(n, &Item{}, v)
			}
		case ebnf.Alternative:
			n.this = x
			join(root, n)
			for _, v := range x {
				dump(n, &Item{}, v)
			}
		case *ebnf.Option:
			n.this = x
			join(root, n)
			dump(n, &Item{}, x.Body)
		case *ebnf.Repetition:
			n.this = x
			join(root, n)
			dump(n, &Item{}, x.Body)
		case *ebnf.Group:
			n.this = x
			join(root, n)
			dump(n, &Item{}, x.Body)
		case *ebnf.Token:
			n.this = x
			join(root, n)
		case *ebnf.Name:
			if cache[x.String] == nil {
				p := g[x.String]
				if p != nil {
					cache[x.String] = p
					passed[x.String] = &Item{}
					dump(root, passed[x.String], p)
				} else {
					cache[x.String] = x
					passed[x.String] = &Item{this: x}
					join(root, passed[x.String])
				}
			} else {
				join(root, passed[x.String])
			}
		default:
			halt.As(100, reflect.TypeOf(x))
		}
	}
	ret := &Item{}
	dump(nil, ret, g)
	return reorder(ret)
}
