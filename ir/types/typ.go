package types

type Type int

const (
	Undef Type = iota
	INTEGER
	BOOLEAN
	TRILEAN
	CHAR
	STRING
	ATOM
	REAL
	// leave this last
	NONE
)

var TypMap map[string]Type

func (t Type) String() (ret string) {
	switch t {
	case INTEGER:
		return "INTEGER"
	case BOOLEAN:
		return "BOOLEAN"
	case TRILEAN:
		return "TRILEAN"
	case CHAR:
		return "CHAR"
	case STRING:
		return "STRING"
	case ATOM:
		return "ATOM"
	case REAL:
		return "REAL"
	default:
		return ""
	}
}

func init() {
	TypMap = make(map[string]Type)
	for i := int(Undef); i < int(NONE); i++ {
		TypMap[Type(i).String()] = Type(i)
	}
}
