package lead

import (
	"leaf/ir"
	"leaf/lss"
)

type Parser interface {
	Import() (*ir.Import, error)
}

var ConnectTo func(s lss.Scanner) Parser
