/*
	lenin - leaf naive interpreter
*/
package lenin

import (
	"github.com/kpmy/ypk/assert"
	"leaf/ir"
)

type Loader func(string) (*ir.Module, error)

var Run func(*ir.Module, Loader)

func Do(m *ir.Module, ld Loader) {
	assert.For(Run != nil, 0)
	assert.For(m != nil, 20)
	Run(m, ld)
}
