package ir

type Object interface {
	Name(string)
}

type Variable struct {
	name string
}

func (v *Variable) Name(s string) { v.name = s }

type Scope struct {
	objects []interface{}
}

func (s *Scope) Add(o Object) {
	s.objects = append(s.objects, o)
}

type Module struct {
	Name string
	Top  *Scope
}

func NewMod(name string) *Module {
	return &Module{Name: name, Top: &Scope{}}
}

func NewVar() *Variable {
	return &Variable{}
}
