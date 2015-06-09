package ir

type Type int

const (
	Undef Type = iota
	Pointer
	Map
	List
	Set
	Integer
	Real
	Boolean
	Trilean
	String
	Atom
	Char
)
