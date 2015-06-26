package tt

import (
	"bufio"
	"fmt"
	"io"
	"leaf/ir"
	"leaf/lead/target"
)

type generator struct {
	wr *bufio.Writer
	m  *ir.Module
}

func (g *generator) sprint(x ...interface{}) {
	g.wr.WriteString(fmt.Sprint(x...))
}

func (g *generator) ln(x ...interface{}) {
	g.wr.WriteString(fmt.Sprint(x...))
	g.wr.WriteString(fmt.Sprintln())
}

func (g *generator) tab(n ...int) {
	if len(n) > 0 {
		for i := 0; i < n[0]; i++ {
			g.wr.WriteRune('\t')
		}
	} else {
		g.wr.WriteRune('\t')
	}
}

func (g *generator) module() {
	g.sprint("DEFINITION ", g.m.Name)
}

func load(sc io.Reader) (ret *ir.Module) {
	panic(0)
}

func store(mod *ir.Module, tg io.Writer) {
	g := &generator{}
	g.m = mod
	g.wr = bufio.NewWriter(tg)
	g.module()
	g.wr.Flush()
}

func init() {
	target.Ext = store
	target.Int = load
}
