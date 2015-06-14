package parser

import (
	"leaf/ir"
)

type target struct {
	root *ir.Module
}

func (t *target) init(mod string) {
	t.root = &ir.Module{ModName: mod}
}
