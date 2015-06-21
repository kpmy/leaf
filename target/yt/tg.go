//Target for compiler, stores AST in yaml
package yt

import (
	"bytes"
	"fmt"
	"github.com/kpmy/ypk/assert"
	"github.com/kpmy/ypk/halt"
	"gopkg.in/yaml.v2"
	"io"
	"leaf/ir"
	"leaf/ir/operation"
	"leaf/ir/types"
	"leaf/target"
	"reflect"
)

type Expression struct {
	Type ExprType
	Leaf map[string]interface{} `yaml:"leaf,omitempty"`
}

type Var struct {
	Guid string
	Type string
}

type Const struct {
	Guid string
	Expr *Expression `yaml:"expression"`
}

type Statement struct {
	Type StmtType               `yaml:"statement"`
	Leaf map[string]interface{} `yaml:"leaf,omitempty"`
}

type Module struct {
	Name      string
	ConstDecl map[string]*Const `yaml:"const,omitempty"`
	VarDecl   map[string]*Var   `yaml:"var,omitempty"`
	BeginSeq  []*Statement      `yaml:"begin,omitempty"`
	CloseSeq  []*Statement      `yaml:"close,omitempty"`

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

type futureThat func() interface{}

func (m *Module) that(id string, i ...interface{}) (ret interface{}) {
	find := func(s string) (ret interface{}) {
		for k, v := range m.id {
			if v == s {
				ret = k
			}
		}
		return
	}
	if x := find(id); x == nil {
		if len(i) == 1 {
			m.id[i[0]] = id
		} else {
			ret = func() interface{} {
				if x := find(id); x != nil {
					return x
				} else {
					halt.As(100, "undefined object ", id)
				}
				return nil
			}
		}
	} else {
		ret = x
	}
	return
}

func typeFix(e *ir.ConstExpr) {
	switch e.Type {
	case types.INTEGER, types.BOOLEAN, types.TRILEAN, types.CHAR, types.STRING, types.REAL:
		//TODO later
	default:
		halt.As(100, "unknown constant type ", e.Type)
	}
}

func internalize(m *Module) (ret *ir.Module) {
	ret = &ir.Module{}
	ret.Init()
	ret.Name = m.Name
	var expr func(e *Expression) ir.Expression
	expr = func(e *Expression) ir.Expression {
		d := &dumbExpr{}
		switch e.Type {
		case Atom:
			this := &ir.AtomExpr{}
			this.Value = e.Leaf["name"].(string)
			d.later = func() ir.Expression { return this }
		case Constant:
			this := &ir.ConstExpr{}
			this.Value = e.Leaf["value"]
			this.Type = types.TypMap[e.Leaf["type"].(string)]
			typeFix(this)
			d.e = this
		case NamedConstant:
			this := &ir.NamedConstExpr{}
			id := e.Leaf["object"].(string)
			_n := m.that(id)
			if n, ok := _n.(*ir.Const); ok {
				this.Named = n
				d.e = this
			} else if _n != nil {
				d.later = func() ir.Expression {
					fn := _n.(func() interface{})
					if x, ok := fn().(*ir.Const); ok {
						this.Named = x
						return this
					} else {
						halt.As(101, "wrong future expr")
					}
					panic(0)
				}
			} else {
				halt.As(100, "unexpected nil")
			}
		case Variable:
			this := &ir.VariableExpr{}
			id := e.Leaf["object"].(string)
			this.Obj = m.that(id).(*ir.Variable)
			d.e = this
		case Monadic:
			this := &ir.Monadic{}
			this.Operand = expr(treatExpr(e.Leaf["operand"]))
			this.Op = operation.OpMap[e.Leaf["operation"].(string)]
			d.e = this
		case Dyadic:
			this := &ir.Dyadic{}
			this.Left = expr(treatExpr(e.Leaf["left"]))
			this.Right = expr(treatExpr(e.Leaf["right"]))
			this.Op = operation.OpMap[e.Leaf["operation"].(string)]
			d.e = this
		default:
			halt.As(100, "unknown type ", e.Type)
		}
		assert.For(d.e != nil || d.later != nil, 60)
		return d
	}

	{
		for k, v := range m.ConstDecl {
			c := &ir.Const{}
			c.Name = k
			c.Expr = expr(v.Expr)
			m.that(v.Guid, c)
			ret.ConstDecl[k] = c
		}
	}

	{
		for k, v := range m.VarDecl {
			i := &ir.Variable{}
			i.Name = k
			i.Type = types.TypMap[v.Type]
			m.that(v.Guid, i)
			ret.VarDecl[k] = i
		}
	}
	stmt := func(s *Statement) (ret ir.Statement) {
		switch s.Type {
		case Assign:
			this := &ir.AssignStmt{}
			this.Object = m.that(s.Leaf["object"].(string)).(*ir.Variable)
			this.Expr = expr(treatExpr(s.Leaf["expression"]))
			ret = this
		default:
			halt.As(100, "unexpected ", s.Type)
		}
		return
	}
	{
		for _, v := range m.BeginSeq {
			ret.BeginSeq = append(ret.BeginSeq, stmt(v))
		}
		for _, v := range m.CloseSeq {
			ret.CloseSeq = append(ret.CloseSeq, stmt(v))
		}
	}
	return
}

func externalize(mod *ir.Module) (ret *Module) {
	ret = &Module{}
	ret.init()
	ret.Name = mod.Name

	var expr func(ir.Expression) *Expression
	expr = func(_e ir.Expression) (ex *Expression) {
		ex = &Expression{}
		ex.Leaf = make(map[string]interface{})
		switch e := _e.(type) {
		case *ir.AtomExpr:
			ex.Type = Atom
			ex.Leaf["name"] = e.Value
		case *ir.ConstExpr:
			ex.Type = Constant
			ex.Leaf["value"] = e.Value
			ex.Leaf["type"] = e.Type.String()
		case *ir.NamedConstExpr:
			ex.Type = NamedConstant
			ex.Leaf["object"] = ret.this(e.Named)
		case *ir.VariableExpr:
			ex.Type = Variable
			ex.Leaf["object"] = ret.this(e.Obj)
		case *ir.Monadic:
			ex.Type = Monadic
			ex.Leaf["operand"] = expr(e.Operand)
			ex.Leaf["operation"] = e.Op.String()
		case *ir.Dyadic:
			ex.Type = Dyadic
			ex.Leaf["left"] = expr(e.Left)
			ex.Leaf["right"] = expr(e.Right)
			ex.Leaf["operation"] = e.Op.String()
		case *dumbExpr:
			return expr(e.Eval())
		default:
			halt.As(100, "unexpected ", reflect.TypeOf(e))
		}
		return
	}

	{
		for _, _v := range mod.ConstDecl {
			c := &Const{}
			c.Guid = ret.this(_v)
			var e ir.Expression
			switch v := _v.Expr.(type) {
			case ir.EvaluatedExpression:
				e = v.Eval()
			case *ir.AtomExpr:
				e = v
			default:
				halt.As(100, "unknown expression ", reflect.TypeOf(v))
			}
			assert.For(e != nil, 40)
			c.Expr = expr(e)
			ret.ConstDecl[_v.Name] = c
		}
	}
	{
		for _, v := range mod.VarDecl {
			i := &Var{}
			i.Guid = ret.this(v)
			i.Type = v.Type.String()
			ret.VarDecl[v.Name] = i
		}
	}
	stmt := func(_s ir.Statement) (st *Statement) {
		st = &Statement{}
		st.Leaf = make(map[string]interface{})
		switch s := _s.(type) {
		case *ir.AssignStmt:
			st.Type = Assign
			st.Leaf["object"] = ret.this(s.Object)
			e := s.Expr.(ir.EvaluatedExpression).Eval()
			st.Leaf["expression"] = expr(e)
		default:
			halt.As(100, "unexpected ", reflect.TypeOf(s))
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

func store(mod *ir.Module, tg io.Writer) {
	m := externalize(mod)
	if data, err := yaml.Marshal(m); err == nil {
		wrote, err := tg.Write(data)
		if wrote != len(data) || err != nil {
			halt.As(101, err)
		}
	} else {
		halt.As(100, err)
	}
}

func load(sc io.Reader) (ret *ir.Module) {
	buf := bytes.NewBuffer(nil)
	io.Copy(buf, sc)
	m := &Module{}
	m.init()
	if err := yaml.Unmarshal(buf.Bytes(), m); err == nil {
		ret = internalize(m)
	} else {
		halt.As(100, err)
	}
	return
}

func init() {
	target.Ext = store
	target.Int = load
}
