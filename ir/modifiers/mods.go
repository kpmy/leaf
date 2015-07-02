package modifiers

import (
	"leaf/lss"
	"strconv"
)

type Modifier int

const (
	None Modifier = iota
	Semi
	Full
	//leave this last
	NONE
)

var ModMap map[string]Modifier

func (m Modifier) Sym() lss.Symbol {
	switch m {
	case None:
		return lss.None
	case Semi:
		return lss.Minus
	case Full:
		return lss.Plus
	default:
		return lss.None
	}
}

func (m Modifier) String() string {
	switch m {
	case None:
		return "none"
	case Semi:
		return "semi"
	case Full:
		return "full"
	default:
		return strconv.Itoa(int(m))
	}
}

func init() {
	ModMap = make(map[string]Modifier)
	for i := int(None); i < int(NONE); i++ {
		ModMap[Modifier(i).String()] = Modifier(i)
	}
}
