package lenin

import (
	"container/list"
	"fmt"
	"github.com/kpmy/trigo"
	"github.com/kpmy/ypk/assert"
	"github.com/kpmy/ypk/fn"
	"github.com/kpmy/ypk/halt"
	"leaf/ir"
	"leaf/ir/modifiers"
	"leaf/ir/types"
	"leaf/lem"
	"reflect"
)

type later struct {
	x   interface{}
	par []interface{}
}

type anyData interface {
	String() string
	read() interface{}
	write(interface{})
}

type direct struct {
	____x interface{}
}

func (d *direct) String() string {
	return fmt.Sprint(d.____x)
}

func (d *direct) set(x interface{}) {
	_, fake := x.(*value)
	assert.For(!fake, 21)
	d.____x = x
}

func (d *direct) read() interface{} { return d.____x }
func (d *direct) write(x interface{}) {
	assert.For(x != nil, 20)
	d.set(x)
}

type indirect struct {
	sel   ir.Selector
	ctx   *context
	stor  *storage
	____x interface{}
}

func (i *indirect) String() string {
	if i.sel != nil {
		return fmt.Sprint("@", reflect.TypeOf(i.sel))
	} else {
		return fmt.Sprint("_", i.____x)
	}
}
func (d *indirect) set(x interface{}) {
	_, fake := x.(*value)
	assert.For(!fake, 21)
	d.____x = x
}

func (d *indirect) doSel(in, out *value, end func(*value) *value) {
	d.ctx.sel(d.sel, in, out, end)
}

func (d *indirect) read() (ret interface{}) {
	if d.sel != nil {
		//fmt.Println("indirect")
		d.stor.lock = &lock{}
		d.doSel(nil, nil, func(v *value) *value {
			ret = v.val
			return nil
		})
		d.stor.lock = nil
		_, fake := ret.(*value)
		assert.For(!fake, 21)
		return
	} else {
		return d.____x
	}
}

func (d *indirect) write(x interface{}) {
	assert.For(x != nil, 20)
	_, fake := x.(*value)
	assert.For(!fake, 21)
	if d.sel != nil {
		d.stor.lock = &lock{}
		d.doSel(nil, nil, func(v *value) *value {
			return &value{typ: v.typ, val: x}
		})
		d.stor.lock = nil
	} else {
		d.set(x)
	}
}

type storage struct {
	root     *ir.Module
	link     interface{}
	schema   map[string]*ir.Variable
	data     map[string]anyData
	wrappers map[string]func(*value) *value
	lock     *lock
	prev     *storage
}

type lock struct{}

type targetStack struct {
	sl *list.List
}

func (s *targetStack) push(st *storage) {
	assert.For(st != nil, 20)
	s.sl.PushFront(st)
}

func (s *targetStack) pop() (ret *storage) {
	if s.sl.Len() > 0 {
		el := s.sl.Front()
		ret = s.sl.Remove(el).(*storage)
	}
	return
}

/*func (s *targetStack) top() *storage {
	if s.sl.Len() > 0 {
		return s.sl.Front().Value.(*storage)
	}
	return nil
}*/

func (s *targetStack) init(o *context) {
	s.sl = list.New()
}

type storeStack struct {
	store map[string]*storage
	sl    *list.List
	mt    string
	owner *context
}

func (s *storeStack) init(o *context) {
	s.store = make(map[string]*storage)
	s.sl = list.New()
	s.owner = o
}

func (s *storeStack) push(st *storage) {
	assert.For(st != nil, 20)
	st.prev = s.top()
	s.sl.PushFront(st)
}

func (s *storeStack) pop() (ret *storage) {
	if s.sl.Len() > 0 {
		el := s.sl.Front()
		ret = s.sl.Remove(el).(*storage)
	} else {
		halt.As(100, "pop on empty stack")
	}
	return
}

func (s *storeStack) top() *storage {
	if s.sl.Len() > 0 {
		return s.sl.Front().Value.(*storage)
	}
	return nil
}

func (s *storeStack) alloc(_x interface{}) {
	switch x := _x.(type) {
	case *ir.Module:
		d := &storage{root: x}
		d.init()
		d.alloc(x.VarDecl)
		s.store[x.Name] = d
		if lem.Debug {
			fmt.Println("alloc", x.Name, d.data)
		}
	case *ir.Procedure:
		assert.For(s.mt != "", 20)
		d := &storage{root: s.store[s.mt].root, link: x}
		d.init()
		d.alloc(x.VarDecl)
		s.push(d)
		if lem.Debug {
			fmt.Println("alloc", x.Name, d.data)
		}
	case ir.ImportProcedure:
		assert.For(s.mt != "", 20)
		d := &storage{root: s.store[s.mt].root, link: x}
		d.init()
		d.alloc(x.This().VarDecl)
		s.push(d)
		if lem.Debug {
			fmt.Println("alloc", x.Name(), d.data)
		}
	default:
		halt.As(100, reflect.TypeOf(x))
	}
}

func (s *storeStack) dealloc(_x interface{}) (ret *storage) {
	switch x := _x.(type) {
	case *ir.Module:
		if lem.Debug {
			fmt.Println("dealloc", x.Name, s.store[x.Name].data)
		}
		s.store[x.Name] = nil
		//TODO проверить наличие связанных элементов стека
	case *ir.Procedure:
		assert.For(s.top() != nil && s.top().link == x, 20)
		if lem.Debug {
			fmt.Println("dealloc", x.Name, s.top().data)
		}
		ret = s.top()
		s.pop()
	case ir.ImportProcedure:
		assert.For(s.top() != nil && s.top().link == x, 20)
		if lem.Debug {
			fmt.Println("dealloc", x.Name(), s.top().data)
		}
		ret = s.top()
		s.pop()
	default:
		halt.As(100, reflect.TypeOf(x))
	}
	return
}

func (s *storage) init() {
	s.schema = make(map[string]*ir.Variable)
	s.data = make(map[string]anyData)
	s.wrappers = make(map[string]func(*value) *value)
}

func (s *storage) alloc(vl map[string]*ir.Variable) {
	assert.For(vl != nil, 20)
	s.schema = vl
	for _, v := range s.schema {
		s.wrappers[v.Name] = func(v *value) *value { return v }
		init := func(val interface{}) (ret anyData) {
			assert.For(val != nil, 20)
			switch v.Modifier {
			case modifiers.Full:
				x := &indirect{stor: s}
				x.set(val)
				ret = x
			case modifiers.Semi, modifiers.None:
				x := &direct{}
				x.set(val)
				ret = x
			default:
				halt.As(100, "wrong modifier ", v.Modifier)
			}
			return
			panic(0)
		}

		switch v.Type {
		case types.INTEGER:
			s.data[v.Name] = init(NewInt(0))
		case types.BOOLEAN:
			s.data[v.Name] = init(false)
		case types.TRILEAN:
			s.data[v.Name] = init(tri.NIL)
		case types.CHAR:
			s.data[v.Name] = init(rune(0))
		case types.STRING:
			s.data[v.Name] = init("")
		case types.ATOM:
			s.data[v.Name] = init(Atom(""))
		case types.REAL:
			s.data[v.Name] = init(NewRat(0.0))
		case types.COMPLEX:
			s.data[v.Name] = init(NewCmp(0.0, 0.0))
		case types.ANY:
			s.data[v.Name] = init(&Any{})
		case types.LIST:
			s.data[v.Name] = init(&List{})
		case types.SET:
			s.data[v.Name] = init(&Set{})
		case types.MAP:
			s.data[v.Name] = init(&Map{})
		case types.PTR:
			s.data[v.Name] = init(&Ptr{})
		case types.PROC:
			s.data[v.Name] = init(&Proc{})
		default:
			halt.As(100, "unknown type ", v.Name, ": ", v.Type)
		}
	}
}

func (s *storage) List() (ret []*ir.Variable) {
	for _, x := range s.schema {
		ret = append(ret, x)
	}
	return
}

func (s *storage) Get(name string) interface{} {
	if d := s.data[name]; d != nil {
		return d.read()
	} else {
		halt.As(100, "object not found")
	}
	panic(0)
}

func (s *storage) Set(name string, x interface{}) {
	assert.For(!fn.IsNil(x), 20)
	if d := s.data[name]; d != nil {
		d.write(x)
	} else {
		halt.As(100, "object not found")
	}
}

func (st *storeStack) ref(o *ir.Variable, sel ir.Selector) {
	find := func(s *storage) (ret bool) {
		if data, ok := s.data[o.Name]; ok {
			assert.For(data != nil, 20)
			r := data.(*indirect)
			r.sel = sel
			r.ctx = st.owner
			ret = true
		}
		return
	}
	found := false
	if local := st.top(); local != nil {
		found = find(local)
	}
	assert.For(found, 60)
}

func (st *storeStack) find(s *storage, o *ir.Variable, fn func(*value) *value) (ret bool) {
	//fmt.Println(o.Name)
	if data, ok := s.data[o.Name]; ok {
		assert.For(data != nil, 20)
		wr := s.wrappers[o.Name]
		nv := fn(wr(&value{typ: o.Type, val: data.read()}))
		if nv != nil {
			assert.For(compTypes(nv.typ, o.Type), 40, "provided ", nv.typ, " != expected ", o.Type)
			nv = conv(nv, o.Type)
			s.data[o.Name].write(nv.val)
			if lem.Debug {
				fmt.Println("touch", o.Name, nv.val)
			}
		}
		ret = true
	}
	return
}
func (s *storeStack) inner(o *ir.Variable, fn func(*value) *value) {
	found := false
	for local := s.top(); local != nil; {
		if local.lock != nil {
			//fmt.Println("locked, try prev")
			local = local.prev
		} else {
			found = s.find(local, o, fn)
			break
		}
	}
	if !found {
		found = s.find(s.store[s.mt], o, fn)
	}
	assert.For(found, 60, `"`, o.Name, `"`)
}

func (s *storeStack) wrap(id string, fn func(*value) *value) {
	find := func(s *storage) (ret bool) {
		if _, ok := s.data[id]; ok {
			if fn == nil {
				fn = func(v *value) *value { return v }
			}
			s.wrappers[id] = fn
			ret = true
		}
		return
	}
	found := false
	for local := s.top(); local != nil; {
		if local.lock != nil {
			//fmt.Println("locked, try prev")
			local = local.prev
		} else {
			found = find(local)
			break
		}
	}
	if !found {
		found = find(s.store[s.mt])
	}
	assert.For(found, 60, `"`, id, `"`)
}

func (s *storeStack) outer(st *storage, o *ir.Variable, fn func(*value) *value) {
	if st != nil {
		found := s.find(st, o, fn)
		assert.For(found, 60, st.root.Name, ".", o.Name)
	} else {
		s.inner(o, fn)
		return
	}
}

type exprStack struct {
	vl *list.List
}

func (s *exprStack) init() {
	s.vl = list.New()
}

func (s *exprStack) push(v *value) {
	assert.For(v != nil, 20)
	_, fake := v.val.(*value)
	assert.For(!fake, 21)
	s.vl.PushFront(v)
}

func (s *exprStack) pop() (ret *value) {
	if s.vl.Len() > 0 {
		el := s.vl.Front()
		ret = s.vl.Remove(el).(*value)
	} else {
		halt.As(100, "pop on empty stack")
	}
	return
}
