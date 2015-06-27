package tt

import (
	"bufio"
	"fmt"
	"github.com/kpmy/ypk/halt"
	"io"
	"leaf/ir"
	"leaf/ir/modifiers"
	"leaf/lead/target"
	"reflect"
	"sort"
)

type generator struct {
	wr *bufio.Writer
	m  *ir.Module
}

func (g *generator) str(x ...interface{}) {
	g.wr.WriteString(fmt.Sprint(x...))
}

func (g *generator) ln(x ...interface{}) {
	g.wr.WriteString(fmt.Sprint(x...))
	g.wr.WriteString(fmt.Sprintln())
}

func (g *generator) tab(n ...int) {
	if len(n) > 0 {
		for i := 0; i < n[0]; i++ {
			g.wr.WriteRune('\t')
		}
	} else {
		g.wr.WriteRune('\t')
	}
}

func (g *generator) expr(x ir.Expression) {
	buf := ""
	put := func(x ...interface{}) {
		buf = fmt.Sprint(buf, fmt.Sprint(x...))
	}
	var expr func(ir.Expression)
	expr = func(_e ir.Expression) {
		switch e := _e.(type) {
		case ir.EvaluatedExpression:
			put("(")
			expr(e.Eval())
			put(")")
		case *ir.ConstExpr:
			put(e.Value)
		case *ir.Dyadic:
			expr(e.Left)
			put(" ", e.Op.String(), " ")
			expr(e.Right)
		case *ir.NamedConstExpr:
			put(e.Named.Name)
		case *ir.VariableExpr:
			put(e.Obj.Name)
		default:
			halt.As(100, reflect.TypeOf(e))
		}
	}
	expr(x)
	g.str(buf)
}

func (g *generator) cdecl(cd map[string]*ir.Const) {
	var tmp []string
	for _, c := range cd {
		if c.Modifier == modifiers.Full {
			tmp = append(tmp, c.Name)
		}
	}
	sort.Strings(tmp)
	if len(tmp) > 0 {
		g.tab()
		g.ln("CONST")
	}
	for _, cn := range tmp {
		c := cd[cn]
		g.tab(2)
		g.str(cn)
		if _, ok := c.Expr.(*ir.AtomExpr); ok {
			g.ln()
		} else if e, ok := c.Expr.(ir.EvaluatedExpression); ok {
			if _, ok := e.Eval().(*ir.AtomExpr); ok {
				g.ln()
			} else {
				g.str(" = ")
				g.expr(c.Expr)
				g.ln()
			}
		}
	}
}

func (g *generator) vdecl(vd map[string]*ir.Variable) {
	var tmp []string
	for _, v := range vd {
		if v.Modifier != modifiers.None {
			tmp = append(tmp, v.Name)
		}
	}
	if len(tmp) > 0 {
		g.tab()
		g.ln("VAR")
	}
	sort.Strings(tmp)
	for _, vn := range tmp {
		v := vd[vn]
		g.tab(2)
		g.str(v.Name)
		g.str(v.Modifier.Sym().String())
		g.ln(" ", v.Type.String())
	}
}

func (g *generator) pdecl(pd map[string]*ir.Procedure) {
	var tmp []string
	for _, p := range pd {
		if p.Modifier == modifiers.Full {
			tmp = append(tmp, p.Name)
		}
	}
	sort.Strings(tmp)
	for _, pn := range tmp {
		p := pd[pn]
		g.tab()
		g.ln("PROCEDURE", " ", p.Name)
		if len(p.VarDecl) > 0 {
			g.vdecl(p.VarDecl)
			if len(p.Infix) > 0 {
				g.tab()
				g.str("INFIX ")
				for _, i := range p.Infix {
					g.str(i.Name, " ")
				}
			}
			g.ln()
		}
		for _, c := range p.Pre {
			g.tab()
			g.str("PRE ")
			g.expr(c)
			g.ln()
		}
		for _, c := range p.Post {
			g.tab()
			g.str("POST ")
			g.expr(c)
			g.ln()
		}
		g.tab()
		g.ln("END ", p.Name)
		g.ln()
	}
}

func (g *generator) module() {
	g.ln("DEFINITION ", g.m.Name)
	g.ln()
	if len(g.m.ImportSeq) > 0 {
		g.tab()
		g.str("IMPORT")
		for _, i := range g.m.ImportSeq {
			g.str(" ", i.Name)
		}
		g.ln()
		g.ln()
	}
	if len(g.m.ConstDecl) > 0 {
		g.cdecl(g.m.ConstDecl)
	}
	if len(g.m.VarDecl) > 0 {
		g.vdecl(g.m.VarDecl)
	}
	if len(g.m.ProcDecl) > 0 {
		g.ln()
		g.pdecl(g.m.ProcDecl)
	}
	g.str("END ", g.m.Name, ".")
}

func load(sc io.Reader) (ret *ir.Module) {
	panic(0)
}

func store(mod *ir.Module, tg io.Writer) {
	g := &generator{}
	g.m = mod
	g.wr = bufio.NewWriter(tg)
	g.module()
	g.wr.Flush()
}

func init() {
	target.Ext = store
	target.Int = load
}
