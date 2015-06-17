package types

type Type int

const (
	Undef Type = iota
	INTEGER
	BOOLEAN
	TRILEAN
)

func (t Type) String() (ret string) {
	switch t {
	case INTEGER:
		return "INTEGER"
	case BOOLEAN:
		return "BOOLEAN"
	case TRILEAN:
		return "TRILEAN"
	default:
		return ""
	}
}
