package li

import (
	"github.com/kpmy/ypk/assert"
	"leaf/ir"
)

var Run func(*ir.Module)

func Do(m *ir.Module) {
	assert.For(Run != nil, 0)
	assert.For(m != nil, 20)
	Run(m)
}
