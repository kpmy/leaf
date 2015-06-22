package z

import (
	"io"
	"leaf/ir"
	"leaf/target"
)

func load(sc io.Reader) (ret *ir.Module) {
	return nil
}

func store(mod *ir.Module, tg io.Writer) {
}

func init() {
	target.Ext = store
	target.Int = load
}
