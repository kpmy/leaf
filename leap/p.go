package leap

import (
	"github.com/kpmy/ypk/assert"
	"leaf/ir"
	"leaf/ir/modifiers"
	"leaf/ir/types"
	"leaf/lead"
	"leaf/lss"
)

type Type struct {
	typ types.Type
}

var entries map[lss.Foreign]interface{}
var idents map[string]lss.Foreign
var mods map[lss.Symbol]modifiers.Modifier

const (
	none lss.Foreign = iota
	integer
	comp
	flo //for real, real is builtin go function O_o
	char
	str
	atom
	boolean
	trilean
)

func init() {
	idents = map[string]lss.Foreign{"INTEGER": integer,
		"BOOLEAN": boolean,
		"TRILEAN": trilean,
		"CHAR":    char,
		"STRING":  str,
		"ATOM":    atom,
		"REAL":    flo,
		"COMPLEX": comp}

	entries = map[lss.Foreign]interface{}{integer: Type{typ: types.INTEGER},
		boolean: Type{typ: types.BOOLEAN},
		trilean: Type{typ: types.TRILEAN},
		char:    Type{typ: types.CHAR},
		str:     Type{typ: types.STRING},
		atom:    Type{typ: types.ATOM},
		flo:     Type{typ: types.REAL},
		comp:    Type{typ: types.COMPLEX}}

	mods = map[lss.Symbol]modifiers.Modifier{lss.Minus: modifiers.Semi, lss.Plus: modifiers.Full}

	lead.ConnectTo = leadp
}

type Parser interface {
	Module() (*ir.Module, error)
}

type Resolver func(name string) (*ir.Import, error)

type pr struct {
	common
	target
	resolver Resolver
}

func (p *pr) resolve(name string) (ret *ir.Import) {
	ret, _ = p.resolver(name)
	if ret == nil {
		p.mark("unresolved import")
	}
	return
}

func (p *pr) init() {
	for k, v := range idents {
		p.sc.Register(v, k)
	}
	p.next()
}

func (p *pr) constDecl(b *constBuilder) {
	assert.For(p.sym.Code == lss.Const, 20, "CONST block expected")
	p.next()
	for {
		if p.await(lss.Ident, lss.Delimiter, lss.Separator) {
			id := p.ident()
			if c, free := b.this(id); c != nil || !free {
				p.mark("identifier already exists")
			}
			p.next()
			obj := &ir.Const{Name: id}
			if p.await(lss.Plus) {
				obj.Modifier = mods[p.sym.Code]
				p.next()
			}
			if p.await(lss.Equal, lss.Separator) { //const expression
				p.next()
				p.pass(lss.Separator)
				obj.Expr = &exprBuilder{sc: b.sc}
				p.expression(obj.Expr.(*exprBuilder))
			} else if p.is(lss.Delimiter) { //ATOM
				obj.Expr = &ir.AtomExpr{Value: id}
				p.next()
			} else {
				p.mark("delimiter or expression expected")
			}
			b.decl(id, obj)
		} else {
			break
		}
	}
}

func (p *common) typ(consume func(t types.Type)) {
	assert.For(p.sym.Code == lss.Ident, 20, "type identifier expected here")
	id := p.ident()
	if t, ok := entries[p.sym.User].(Type); ok {
		switch t.typ {
		case types.INTEGER, types.REAL, types.COMPLEX:
			p.next()
			consume(t.typ)
		case types.CHAR, types.STRING:
			p.next()
			consume(t.typ)
		case types.ATOM, types.BOOLEAN, types.TRILEAN:
			p.next()
			consume(t.typ)
		default:
			p.mark("unexpected type ", id)
		}
	} else {
		p.mark("unknown type ", id)
	}
}

func (p *pr) varDecl(b *varBuilder) {
	assert.For(p.sym.Code == lss.Var, 20, "VAR block expected")
	p.next()
	for {
		if p.await(lss.Ident, lss.Delimiter, lss.Separator) {
			var vl []*ir.Variable
			for {
				obj := &ir.Variable{}
				id := p.ident()
				if v, free := b.this(id); v != nil || !free {
					p.mark("identifier already exists")
				}
				obj.Name = id
				vl = append(vl, obj)
				b.decl(obj.Name, obj)
				p.next()
				if p.await(lss.Minus) || p.is(lss.Plus) {
					obj.Modifier = mods[p.sym.Code]
					p.next()
				}
				if p.await(lss.Comma, lss.Separator) {
					p.next()
					p.pass(lss.Separator)
				} else {
					break
				}
			}
			p.expect(lss.Ident, "type identifier expected", lss.Separator)
			p.typ(func(t types.Type) {
				for _, obj := range vl {
					obj.Type = t
				}
			})
		} else {
			break
		}
	}
}

func (p *pr) qualSel(b *blockBuilder) (sel *selBuilder, mid, id string) {
	mod := false
	mid, mod = p.qualident(b.sc)
	id = ""
	if mod {
		sel = &selBuilder{sc: b.sc}
		imp := b.sc.im[mid]
		sel.join(&ir.SelectMod{Mod: imp.Name})
		id = p.ident()
		obj := b.impObj(mid, id)
		if obj != nil {
			sel.join(obj)
		} else {
			sel = nil
		}
		p.next()
	} else {
		sel = &selBuilder{sc: b.sc}
		id = mid
		mid = ""
		obj := b.obj(id)
		if obj != nil {
			sel.join(obj)
		} else {
			sel = nil
		}
	}
	return
}

func (p *pr) stmtSeq(b *blockBuilder) {
	for stop := false; !stop; {
		p.pass(lss.Separator, lss.Delimiter)
		switch p.sym.Code {
		case lss.Ident:
			sel, cm, id := p.qualSel(b)
			if sel != nil {
				p.pass(lss.Separator)
				p.selector(sel)
				if p.is(lss.Becomes) {
					stmt := &ir.AssignStmt{}
					p.next()
					p.pass(lss.Separator)
					expr := &exprBuilder{sc: b.sc}
					p.expression(expr)
					stmt.Sel = sel
					stmt.Expr = expr
					b.put(stmt)
					//p.expect(lss.Delimiter, "delimiter expected", lss.Separator)
				} else {
					p.mark("illegal statement ", p.sym.Code)
				}
			} else {
				var param []*forwardParam
				if p.await(lss.Lparen, lss.Separator, lss.Delimiter) {
					p.next()
					for {
						p.expect(lss.Ident, "identifier expected", lss.Separator, lss.Delimiter)
						id := p.ident()
						p.next()
						if p.await(lss.Colon, lss.Separator) {
							par := &forwardParam{name: id}
							p.next()
							e := &exprBuilder{sc: b.sc}
							p.expression(e)
							par.expr = e
							param = append(param, par)
						} else if p.is(lss.Square) {
							par := &forwardParam{name: id}
							p.next()
							p.expect(lss.Ident, "ident expected", lss.Separator)
							sel, pm, _ := p.qualSel(b)
							if sel == nil {
								p.mark("not an object")
							}
							p.pass(lss.Separator)
							p.selector(sel)
							if pm == "" && cm != "" {
								msel := &ir.SelectMod{Mod: p.target.top.Name}
								sel.head(msel)
							}
							par.link = sel
							param = append(param, par)
						} else {
							p.mark("colon expected")
						}
						if p.await(lss.Comma, lss.Separator, lss.Delimiter) {
							p.next()
						} else {
							break
						}
					}
					p.expect(lss.Rparen, "no ) found", lss.Separator, lss.Delimiter)
					p.next()
				}
				stmt := b.call(cm, id, param)
				b.put(stmt)
			}
		case lss.If:
			stmt := &ir.IfStmt{}
			first := true
			for stop := false; !stop; {
				switch p.sym.Code {
				case lss.If, lss.Elsif:
					if p.is(lss.If) && !first {
						p.mark("ELSIF expected")
					}
					first = false
					p.next()
					p.pass(lss.Separator)
					expr := &exprBuilder{sc: b.sc}
					p.expression(expr)
					p.expect(lss.Then, "THEN not found", lss.Separator)
					p.next()
					st := &blockBuilder{sc: b.sc}
					p.stmtSeq(st)
					p.pass(lss.Separator, lss.Delimiter)
					br := &ir.ConditionBranch{}
					br.Expr = expr
					br.Seq = st.seq
					stmt.Cond = append(stmt.Cond, br)
				case lss.Else:
					p.next()
					st := &blockBuilder{sc: b.sc}
					p.stmtSeq(st)
					p.pass(lss.Separator, lss.Delimiter)
					br := &ir.ElseBranch{}
					br.Seq = st.seq
					stmt.Else = br
				case lss.End:
					p.next()
					stop = true
				default:
					p.mark("END or ELSE/ELSIF expected, but ", p.sym)
				}
			}
			b.put(stmt)
		case lss.While:
			stmt := &ir.WhileStmt{}
			first := true
			for stop := false; !stop; {
				switch p.sym.Code {
				case lss.While, lss.Elsif:
					if p.is(lss.While) && !first {
						p.mark("ELSIF expected")
					}
					first = false
					p.next()
					p.pass(lss.Separator)
					expr := &exprBuilder{sc: b.sc}
					p.expression(expr)
					p.expect(lss.Do, "DO not found", lss.Separator)
					p.next()
					st := &blockBuilder{sc: b.sc}
					p.stmtSeq(st)
					p.pass(lss.Separator, lss.Delimiter)
					br := &ir.ConditionBranch{}
					br.Expr = expr
					br.Seq = st.seq
					stmt.Cond = append(stmt.Cond, br)
				case lss.End:
					p.next()
					stop = true
				default:
					p.mark("END or ELSIF expected, but ", p.sym)
				}
			}
			b.put(stmt)
		case lss.Repeat:
			p.next()
			stmt := &ir.RepeatStmt{}
			br := &ir.ConditionBranch{}
			st := &blockBuilder{sc: b.sc}
			p.pass(lss.Separator, lss.Delimiter)
			p.stmtSeq(st)
			p.expect(lss.Until, "UNTIL expected", lss.Separator, lss.Delimiter)
			p.next()
			expr := &exprBuilder{sc: b.sc}
			p.expression(expr)
			p.expect(lss.Delimiter, "delimiter expected", lss.Separator)
			p.next()
			br.Expr = expr
			br.Seq = st.seq
			stmt.Cond = br
			b.put(stmt)
		case lss.Choose:
			p.next()
			stmt := &ir.ChooseStmt{}
			if !p.await(lss.Of, lss.Separator, lss.Delimiter) {
				expr := &exprBuilder{sc: b.sc}
				p.expression(expr)
				stmt.Expr = expr
				p.next()
				p.expect(lss.Of, "OF expected", lss.Separator, lss.Delimiter)
			}
			first := true
			for stop := false; !stop; {
				switch p.sym.Code {
				case lss.Of, lss.Opt:
					if p.is(lss.Of) && !first {
						p.mark("ELSIF expected")
					}
					first = false
					p.next()
					p.pass(lss.Separator, lss.Delimiter)
					var bunch []ir.Expression
					for {
						expr := &exprBuilder{sc: b.sc}
						p.expression(expr)
						bunch = append(bunch, expr)
						if p.await(lss.Colon, lss.Separator) {
							p.next()
							break
						} else if !p.is(lss.Comma) {
							p.mark("comma expected")
						} else {
							p.next()
						}
					}
					st := &blockBuilder{sc: b.sc}
					p.stmtSeq(st)
					p.pass(lss.Separator, lss.Delimiter)
					for _, e := range bunch {
						br := &ir.ConditionBranch{}
						br.Expr = e
						br.Seq = st.seq
						stmt.Cond = append(stmt.Cond, br)
					}
				case lss.Else:
					p.next()
					st := &blockBuilder{sc: b.sc}
					p.stmtSeq(st)
					p.pass(lss.Separator, lss.Delimiter)
					br := &ir.ElseBranch{}
					br.Seq = st.seq
					stmt.Else = br
				case lss.End:
					stop = true
					p.next()
				default:
					p.mark("END expected")
				}
			}
			b.put(stmt)
		default:
			stop = true
		}
	}
}

func (p *pr) procDecl(b *blockBuilder) {
	assert.For(p.is(lss.Proc), 20, "PROCEDURE expected here")
	ret := &ir.Procedure{}
	ret.Init()
	p.next()
	p.expect(lss.Ident, "procedure name expected", lss.Separator)
	ret.Name = p.ident()
	p.next()
	if p.await(lss.Plus) {
		ret.Modifier = mods[p.sym.Code]
		p.next()
	}
	p.st.push()
	this := p.st.this()
	p.block(this, lss.Proc)
	p.expect(lss.Begin, "BEGIN expected", lss.Separator, lss.Delimiter)
	p.next()
	b.decl(ret.Name, ret)
	proc := &blockBuilder{sc: this}
	p.stmtSeq(proc)
	{
		ret.Seq = proc.seq
		ret.ConstDecl = this.cm
		ret.VarDecl = this.vm
		ret.ProcDecl = this.pm
		ret.Infix = this.in
		ret.Pre = this.pre
		ret.Post = this.post
		expect := modifiers.Full
		for i, v := range ret.Infix {
			if v.Modifier != expect {
				p.mark("wrong infix declared")
			}
			if i == 0 {
				expect = modifiers.Semi
			}
		}
	}
	p.expect(lss.End, "no END", lss.Delimiter, lss.Separator)
	p.next()
	p.expect(lss.Ident, "procedure name expected", lss.Separator)
	if p.ident() != ret.Name {
		p.mark("procedure name does not match")
	}
	p.st.pop()
	p.next()
}

func (p *pr) block(bl *block, typ lss.Symbol) {
	assert.For(typ == lss.Module || typ == lss.Proc, 20, "unknown block type ", typ)
	if typ == lss.Module {
		cache := make(map[string]*ir.Import)
		for p.await(lss.Import, lss.Delimiter, lss.Separator) {
			p.next()
			for stop := false; !stop; {
				if p.await(lss.Ident, lss.Delimiter, lss.Separator) {
					name := ""
					alias := ""
					id := p.ident()
					p.next()
					if p.await(lss.Becomes, lss.Separator) {
						p.next()
						p.expect(lss.Ident, "ident expected", lss.Separator)
						name = p.ident()
						alias = id
						p.next()
					} else {
						name = id
						alias = id
					}
					var i *ir.Import
					if i = bl.im[alias]; i == nil {
						if i = cache[name]; i == nil {
							i = p.resolve(name)
							cache[name] = i
							bl.il = append(bl.il, i)
						}
						bl.im[alias] = i
					} else {
						p.mark("import module already exists")
					}
				} else if len(cache) > 0 {
					stop = true
				} else {
					p.mark("nothing to import")
				}
			}
		}
	}
	for p.await(lss.Const, lss.Delimiter, lss.Separator) {
		b := &constBuilder{sc: bl}
		p.constDecl(b)
	}
	for p.await(lss.Var, lss.Delimiter, lss.Separator) {
		b := &varBuilder{sc: bl}
		p.varDecl(b)
	}
	if typ == lss.Proc { //infix description, pre- and post-conditions, etc
		for stop := false; !stop; {
			p.pass(lss.Delimiter, lss.Separator)
			switch p.sym.Code {
			case lss.Infix:
				p.next()
				for stop := false; !stop; {
					if p.await(lss.Ident, lss.Separator) {
						obj := bl.vm[p.ident()]
						if obj == nil {
							p.mark("unknown identifier")
						}
						bl.in = append(bl.in, obj)
						p.next()
						if p.await(lss.Delimiter, lss.Separator) {
							p.next()
							stop = true
						}
					} else if p.is(lss.Delimiter) {
						stop = true
						p.next()
					} else {
						p.mark("identifier expected", p.sym.Code)
					}
				}
			case lss.Pre:
				p.next()
				expr := &exprBuilder{sc: bl}
				p.expression(expr)
				bl.pre = append(bl.pre, expr)
				p.expect(lss.Delimiter, "delimiter expected", lss.Separator)
			case lss.Post:
				p.next()
				expr := &exprBuilder{sc: bl}
				p.expression(expr)
				bl.post = append(bl.post, expr)
				p.expect(lss.Delimiter, "delimiter expected", lss.Separator)
			default:
				stop = true
			}
		}
	}
	for p.await(lss.Proc, lss.Delimiter, lss.Separator) {
		b := &blockBuilder{sc: bl}
		p.procDecl(b)
	}
}

func (p *pr) Module() (ret *ir.Module, err error) {
	if !p.await(lss.Module, lss.Delimiter, lss.Separator) {
		if p.sc.Error() != nil {
			return nil, p.sc.Error()
		} else {
			p.mark("MODULE expected")
		}
	}
	p.next()
	p.expect(lss.Ident, "module name expected", lss.Separator)
	p.target.init(p.ident())
	p.next()
	p.pass(lss.Separator, lss.Delimiter)
	p.st.push()
	top := p.st.this()
	p.block(top, lss.Module)
	p.st.pop()
	p.top.ConstDecl = top.cm
	p.top.VarDecl = top.vm
	p.top.ProcDecl = top.pm
	p.top.ImportSeq = top.il
	if p.await(lss.Begin, lss.Delimiter, lss.Separator) {
		p.next()
		b := &blockBuilder{sc: top}
		p.stmtSeq(b)
		p.top.BeginSeq = b.seq
	}
	if p.await(lss.Close, lss.Delimiter, lss.Separator) {
		p.next()
		b := &blockBuilder{sc: top}
		p.stmtSeq(b)
		p.top.CloseSeq = b.seq
	}
	//p.run(lss.End)
	p.expect(lss.End, "no END", lss.Delimiter, lss.Separator)
	p.next()
	p.expect(lss.Ident, "module name expected", lss.Separator)
	if p.ident() != p.top.Name {
		p.mark("module name does not match")
	}
	p.next()
	p.expect(lss.Period, "end of module expected")
	ret = p.top
	return
}

func ConnectTo(s lss.Scanner, rs Resolver) Parser {
	assert.For(s != nil, 20)
	s.Init(lss.Module, lss.End, lss.Do, lss.While, lss.Elsif, lss.Import, lss.Const, lss.Of, lss.Pre, lss.Post, lss.Proc, lss.Var, lss.Begin, lss.Close, lss.If, lss.Then, lss.Repeat, lss.Until, lss.Else, lss.True, lss.False, lss.Nil, lss.Inf, lss.Choose, lss.Opt, lss.Infix)
	ret := &pr{resolver: rs}
	ret.sc = s
	ret.debug = false
	ret.init()
	return ret
}
