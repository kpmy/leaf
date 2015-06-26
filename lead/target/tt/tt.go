package tt

import (
	"io"
	"leaf/ir"
	"leaf/lead/target"
)

func load(sc io.Reader) (ret *ir.Module) {
	panic(0)
}

func store(mod *ir.Module, tg io.Writer) {

}

func init() {
	target.Ext = store
	target.Int = load
}
