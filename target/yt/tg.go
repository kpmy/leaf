//Target for compiler, stores AST in yaml
package yt

import (
	"fmt"
	"github.com/kpmy/ypk/halt"
	"gopkg.in/yaml.v2"
	"io"
	"leaf/ir"
	"leaf/target"
	"reflect"
)

type Expression struct {
	Type int
	Leaf map[string]interface{}
}

type Var struct {
	Guid string
	Type int
}

type Const struct {
	Guid string
	Expr *Expression `yaml:"expression"`
}

type Statement struct {
	Type int
	Leaf map[string]interface{}
}

type Module struct {
	Name      string
	ConstDecl map[string]*Const `yaml:"const"`
	VarDecl   map[string]*Var   `yaml:"var"`
	BeginSeq  []*Statement      `yaml:"begin"`
	CloseSeq  []*Statement      `yaml:"close"`

	id map[interface{}]string
}

func (m *Module) init() {
	m.id = make(map[interface{}]string)
	m.ConstDecl = make(map[string]*Const)
	m.VarDecl = make(map[string]*Var)
}

func (m *Module) this(item interface{}) (ret string) {
	if ret = m.id[item]; ret == "" {
		ret = fmt.Sprintf("%X", len(m.id))
		m.id[item] = ret
	}
	return
}

func export(mod *ir.Module) (ret *Module) {
	ret = &Module{}
	ret.init()
	ret.Name = mod.Name

	var expr func(ir.Expression) *Expression
	expr = func(_e ir.Expression) (ex *Expression) {
		ex = &Expression{}
		ex.Leaf = make(map[string]interface{})
		switch e := _e.(type) {
		case *ir.ConstExpr:
			ex.Type = 1
			ex.Leaf["value"] = e.Value
		case *ir.NamedConstExpr:
			ex.Type = 2
			ex.Leaf["object"] = ret.this(e.Named)
		case *ir.VariableExpr:
			ex.Type = 3
			ex.Leaf["object"] = ret.this(e.Obj)
		case *ir.Monadic:
			ex.Type = 4
			ex.Leaf["operand"] = expr(e.Operand)
		case *ir.Dyadic:
			ex.Type = 5
			ex.Leaf["left"] = expr(e.Left)
			ex.Leaf["right"] = expr(e.Right)
		default:
			halt.As(100, "unexpected ", reflect.TypeOf(e))
		}
		return
	}

	{
		for _, v := range mod.ConstDecl {
			c := &Const{}
			c.Guid = ret.this(v)
			e := v.Expr.(ir.EvaluatedExpression).Eval()
			c.Expr = expr(e)
			ret.ConstDecl[v.Name] = c
		}
	}
	{
		for _, v := range mod.VarDecl {
			i := &Var{}
			i.Guid = ret.this(v)
			i.Type = int(v.Type)
			ret.VarDecl[v.Name] = i
		}
	}
	stmt := func(_s ir.Statement) (st *Statement) {
		st = &Statement{}
		st.Leaf = make(map[string]interface{})
		switch s := _s.(type) {
		case *ir.AssignStmt:
			st.Type = 1
			st.Leaf["object"] = ret.this(s.Object)
			e := s.Expr.(ir.EvaluatedExpression).Eval()
			st.Leaf["expression"] = expr(e)
		}
		return
	}
	{
		for _, v := range mod.BeginSeq {
			ret.BeginSeq = append(ret.BeginSeq, stmt(v))
		}
		for _, v := range mod.CloseSeq {
			ret.CloseSeq = append(ret.CloseSeq, stmt(v))
		}
	}
	return
}

func code(mod *ir.Module, tg io.Writer) {
	m := export(mod)
	if data, err := yaml.Marshal(m); err == nil {
		wrote, err := tg.Write(data)
		if wrote != len(data) || err != nil {
			halt.As(101, err)
		}
	} else {
		halt.As(100, err)
	}
}

func init() {
	target.Code = code
}
