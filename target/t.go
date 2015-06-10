package target

type Class int

const (
	Wrong Class = iota
	Variable
)

type Target interface {
	Open(string)
	BeginObject(Class)
	Name(string)
	EndObject()
	Close(string)
}
