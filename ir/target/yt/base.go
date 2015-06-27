//Target for compiler, stores AST in yaml
package yt

import (
	"bytes"
	"fmt"
	"github.com/kpmy/ypk/halt"
	"gopkg.in/yaml.v2"
	"io"
	"leaf/ir"
	"leaf/ir/types"
)

type Param struct {
	Uuid string
	Expr *Expression `yaml:"expression"`
	Sel  []*Selector `yaml:"selector"`
}

type Expression struct {
	Type ExprType
	Leaf map[string]interface{} `yaml:"leaf,omitempty"`
}

type Var struct {
	Uuid     string
	Type     string `yaml:"type,omitempty"`
	Modifier string `yaml:"modifier,omitempty"`
}

type Const struct {
	Uuid     string
	Modifier string      `yaml:"modifier,omitempty"`
	Expr     *Expression `yaml:"expression,omitempty"`
}

type Proc struct {
	Uuid      string
	Modifier  string            `yaml:"modifier,omitempty"`
	ConstDecl map[string]*Const `yaml:"constant,omitempty"`
	VarDecl   map[string]*Var   `yaml:"variable,omitempty"`
	ProcDecl  map[string]*Proc  `yaml:"procedure,omitempty"`
	Infix     []string          `yaml:"infix,omitempty"`
	Seq       []*Statement      `yaml:"sequence,omitempty"`
	Pre       []*Expression     `yaml:"precondition,omitempty"`
	Post      []*Expression     `yaml:"postcondition,omitempty"`
}

type Selector struct {
	Type SelType
	Leaf map[string]interface{} `yaml:"leaf,omitempty"`
}

type Statement struct {
	Type StmtType               `yaml:"statement"`
	Leaf map[string]interface{} `yaml:"leaf,omitempty"`
}

type Condition struct {
	Expr *Expression  `yaml:"expression"`
	Seq  []*Statement `yaml:"block,omitempty"`
}

type Module struct {
	Name      string
	ConstDecl map[string]*Const `yaml:"constant,omitempty"`
	VarDecl   map[string]*Var   `yaml:"variable,omitempty"`
	ProcDecl  map[string]*Proc  `yaml:"procedure,omitempty"`
	BeginSeq  []*Statement      `yaml:"begin,omitempty"`
	CloseSeq  []*Statement      `yaml:"close,omitempty"`
	ImpSeq    []*Import         `yaml:"import,omitempty"`
	id        map[interface{}]string
}

type Import struct {
	Name      string
	ConstDecl map[string]*Const `yaml:"constant,omitempty"`
	VarDecl   map[string]*Var   `yaml:"variable,omitempty"`
	ProcDecl  map[string]*Proc  `yaml:"procedure,omitempty"`
}

func (m *Module) init() {
	m.id = make(map[interface{}]string)
	m.ConstDecl = make(map[string]*Const)
	m.VarDecl = make(map[string]*Var)
	m.ProcDecl = make(map[string]*Proc)
}

func (i *Import) init() {
	i.ConstDecl = make(map[string]*Const)
	i.VarDecl = make(map[string]*Var)
	i.ProcDecl = make(map[string]*Proc)
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

func Store(mod *ir.Module, tg io.Writer) {
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

func Load(sc io.Reader) (ret *ir.Module) {
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
