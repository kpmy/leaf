package modifiers

import (
	"strconv"
)

type Modifier int

const (
	None Modifier = iota
	Semi

	//leave this last
	NONE
)

var ModMap map[string]Modifier

func (m Modifier) String() string {
	switch m {
	case None:
		return "none"
	case Semi:
		return "semi"
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
