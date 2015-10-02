package leap

import (
	"github.com/kpmy/leaf/ir"
	"github.com/kpmy/leaf/leac/lss"
)

type DefResolver func(name string) (*ir.Import, error)

type DefParser interface {
	Import() (*ir.Import, error)
}

var ConnectToDef func(lss.Scanner, DefResolver) DefParser

type ModParser interface {
	Module() (*ir.Module, error)
}

var ConnectToMod func(s lss.Scanner, rs DefResolver) ModParser
