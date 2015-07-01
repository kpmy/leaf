package ym

import (
	"bufio"
	"io"
	"leaf/ir"
	"leaf/lem"
)

type generator struct {
	wr *bufio.Writer
	m  *ir.Module
}

func (g *generator) module() {
}

func load(sc io.Reader) (ret *ir.Module) {
	panic(0)
}

func store(mod *ir.Module, tg io.Writer) {
	g := &generator{}
	g.m = mod
	g.wr = bufio.NewWriter(tg)
	g.module()
}

func init() {
	lem.Ext = store
	lem.Int = load
}
