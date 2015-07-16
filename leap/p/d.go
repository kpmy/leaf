package p

import (
	"github.com/kpmy/ypk/assert"
	"leaf/ir"
	"leaf/ir/modifiers"
	"leaf/ir/types"
	"leaf/leap"
	"leaf/leap/lss"
)

type pd struct {
	common
	imported
	resolver leap.DefResolver
}

func (p *pd) init() {
	for k, v := range idents {
		p.sc.Register(v, k)
	}
	p.next()
}

type ic struct {
	this *ir.Const
}

func (i *ic) Name() string { return i.this.Name }

func (i *ic) Expr() ir.Expression { return i.this.Expr }

func (i *ic) This() *ir.Const { return i.this }

type iv struct {
	this *ir.Variable
}

func (i *iv) Name() string { return i.this.Name }

func (i *iv) Type() types.Type { return i.this.Type }

func (i *iv) Modifier() modifiers.Modifier { return i.this.Modifier }

func (i *iv) This() *ir.Variable { return i.this }

type ip struct {
	this *ir.Procedure
}

func (i *ip) Name() string { return i.this.Name }

func (i *ip) VarDecl() map[string]ir.ImportVariable {
	vm := make(map[string]ir.ImportVariable)
	for k, v := range i.this.VarDecl {
		vm[k] = &iv{this: v}
	}
	return vm
}

func (i *ip) Infix() []ir.ImportVariable {
	var vl []ir.ImportVariable
	for _, v := range i.this.Infix {
		vl = append(vl, &iv{this: v})
	}
	return vl
}

func (i *ip) Pre() []ir.Expression  { return i.this.Pre }
func (i *ip) Post() []ir.Expression { return i.this.Post }
func (i *ip) This() *ir.Procedure   { return i.this }

func (p *pd) resolve(name string) (ret *ir.Import) {
	ret, _ = p.resolver(name)
	if ret == nil {
		p.mark("unresolved import")
	}
	return
}

func (p *pd) constDecl(b *constBuilder) {
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

func (p *pd) varDecl(b *varBuilder) {
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
			if p.await(lss.Ident, lss.Separator) {
				p.typ(func(t types.Type) {
					for _, obj := range vl {
						obj.Type = t
					}
				})
			} else if p.is(lss.Proc) {
				p.next()
				for _, obj := range vl {
					obj.Type = types.PROC
				}
			} else {
				p.mark("identifier expected")
			}
		} else {
			break
		}
	}
}

func (p *pd) procDecl(b *blockBuilder) {
	assert.For(p.is(lss.Proc), 20, "PROCEDURE expected here")
	ret := &ir.Procedure{}
	ret.Init(p.top.Name)
	p.next()
	p.expect(lss.Ident, "procedure name expected", lss.Separator)
	ret.Name = p.ident()
	p.next()
	if p.await(lss.Plus) {
		ret.Modifier = mods[p.sym.Code]
		p.next()
	}
	this := &block{}
	this.init()
	p.block(this, lss.Proc)
	b.decl(ret.Name, ret)
	{
		ret.VarDecl = this.vm
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
	p.next()
}

func (p *pd) block(bl *block, typ lss.Symbol) {
	assert.For(typ == lss.Definition || typ == lss.Proc, 20, "unknown block type ", typ)
	if typ == lss.Definition {
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
	} else {
		for p.await(lss.Proc, lss.Delimiter, lss.Separator) {
			b := &blockBuilder{sc: bl}
			p.procDecl(b)
		}
	}
}

func (p *pd) Import() (*ir.Import, error) {
	if !p.await(lss.Definition, lss.Delimiter, lss.Separator) {
		if p.sc.Error() != nil {
			return nil, p.sc.Error()
		} else {
			p.mark("DEFINITION expected")
		}
	}
	p.next()
	p.expect(lss.Ident, "module name expected", lss.Separator)
	p.imported.init(p.ident())
	p.next()
	p.pass(lss.Separator, lss.Delimiter)

	top := &block{}
	top.init()
	p.block(top, lss.Definition)
	for k, v := range top.cm {
		p.top.ConstDecl[k] = &ic{this: v}
	}
	for k, v := range top.vm {
		p.top.VarDecl[k] = &iv{this: v}
	}
	for k, v := range top.pm {
		p.top.ProcDecl[k] = &ip{this: v}
	}
	p.top.ImportSeq = top.il
	p.expect(lss.End, "no END", lss.Delimiter, lss.Separator)
	p.next()
	p.expect(lss.Ident, "module name expected", lss.Separator)
	if p.ident() != p.top.Name {
		p.mark("module name does not match")
	}
	p.next()
	p.expect(lss.Period, "end of module expected")
	return p.top, nil
}

func leadp(s lss.Scanner, rs leap.DefResolver) leap.DefParser {
	assert.For(s != nil, 20)
	s.Init(lss.Definition, lss.End, lss.Import, lss.Const, lss.Pre, lss.Post, lss.Proc, lss.Var, lss.True, lss.False, lss.Null, lss.Nil, lss.Inf, lss.Choose, lss.Opt, lss.Infix, lss.Undef, lss.Is, lss.In, lss.Rbrux, lss.Lbrux)
	ret := &pd{resolver: rs}
	ret.sc = s
	ret.init()
	return ret
}
