package lead

import (
	"leaf/ir"
	"leaf/lss"
)

type Resolver func(name string) (*ir.Import, error)

type Parser interface {
	Import() (*ir.Import, error)
}

var ConnectTo func(lss.Scanner, Resolver) Parser
