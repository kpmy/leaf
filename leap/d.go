package leap

import (
	"fmt"
	"github.com/kpmy/ypk/assert"
	"leaf/ir"
	"leaf/ir/modifiers"
	"leaf/ir/types"
	"leaf/lead"
	"leaf/lss"
)

type pd struct {
	common
	imported
}

func (p *pd) init() {
	for k, v := range idents {
		p.sc.Register(v, k)
	}
	p.next()
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

func (p *pd) procDecl(b *blockBuilder) {
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
	p.top.ConstDecl = top.cm
	p.top.VarDecl = top.vm
	p.top.ProcDecl = top.pm
	fmt.Println(top.cm, top.vm, top.pm)
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

func leadp(s lss.Scanner) lead.Parser {
	assert.For(s != nil, 20)
	s.Init(lss.Definition, lss.End, lss.Import, lss.Const, lss.Pre, lss.Post, lss.Proc, lss.Var, lss.True, lss.False, lss.Nil, lss.Inf, lss.Choose, lss.Opt, lss.Infix)
	ret := &pd{}
	ret.sc = s
	ret.init()
	return ret
}
