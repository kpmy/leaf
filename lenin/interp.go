/*
	lenin - leaf naive interpreter
*/
package lenin

import (
	"github.com/kpmy/ypk/assert"
	"leaf/ir"
	"leaf/lenin/rt"
)

type Loader func(string) (*ir.Module, error)

var Run func(*ir.Module, Loader, chan rt.Message)

func Do(m *ir.Module, ld Loader, universe chan rt.Message) {
	assert.For(Run != nil, 0)
	assert.For(m != nil, 20)
	Run(m, ld, universe)
}
