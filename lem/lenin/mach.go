package lenin

import (
	"bytes"
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"github.com/kpmy/ypk/halt"
	"io"
	"leaf/ir"
	"leaf/ir/types"
	"leaf/lem"
	"os"
	"path/filepath"
	"reflect"
	"time"
)

const STORAGE = ".store"

type Type int

const (
	Wrong Type = iota
	Machine
	Console
	Kernel
	Storage
)

var TypMap map[string]Type

type unraw struct {
	x *Any
}

type kv struct {
	k, v *unraw
}

func (x *kv) UnmarshalXML(d *xml.Decoder, start xml.StartElement) (err error) {
	var _t xml.Token
	for stop := false; !stop && err == nil; {
		_t, err = d.Token()
		switch tok := _t.(type) {
		case xml.StartElement:
			var z *unraw
			switch tok.Name.Local {
			case "key":
				x.k = &unraw{}
				z = x.k
			case "value":
				x.v = &unraw{}
				z = x.v
			default:
				halt.As(100, tok.Name)
			}
			err = d.DecodeElement(z, &tok)
		case xml.EndElement:
			stop = tok.Name == start.Name
		default:
			halt.As(100, reflect.TypeOf(tok), tok)
		}
	}
	return err
}

func (r *unraw) Value() interface{} { return r.x }

func (r *unraw) UnmarshalXML(d *xml.Decoder, start xml.StartElement) (err error) {
	var t types.Type
	for _, a := range start.Attr {
		switch a.Name.Local {
		case "type":
			t = types.TypMap[a.Value]
		default:
			halt.As(100, a.Name)
		}
	}
	switch t {
	case types.MAP:
		m := &Map{}
		var _t xml.Token
		for stop := false; !stop && err == nil; {
			_t, err = d.Token()
			switch tok := _t.(type) {
			case xml.StartElement:
				k := &kv{}
				err = d.DecodeElement(k, &tok)
				m.Set(k.k.x, k.v.x)
			case xml.EndElement:
				stop = tok.Name == start.Name
			default:
				halt.As(100, reflect.TypeOf(tok), tok)
			}
		}
		r.x = NewAny(t, m)
	case types.STRING:
		var sd xml.Token
		sd, err = d.Token()
		x := string(sd.(xml.CharData))
		r.x = NewAny(t, x)
		_, err = d.Token()
	case types.LIST:
		l := &List{}
		var _t xml.Token
		for stop := false; !stop && err == nil; {
			_t, err = d.Token()
			switch tok := _t.(type) {
			case xml.StartElement:
				i := &unraw{}
				err = d.DecodeElement(i, &tok)
				l.Len(l.Len() + 1)
				l.SetVal(l.Len()-1, i.x)
			case xml.EndElement:
				stop = tok.Name == start.Name
			default:
				halt.As(100, reflect.TypeOf(tok), tok)
			}
		}
		r.x = NewAny(t, l)
	case types.REAL:
		var sd xml.Token
		sd, err = d.Token()
		rr := &Rat{}
		rr.UnmarshalText(sd.(xml.CharData))
		r.x = NewAny(t, rr)
		_, err = d.Token()
	default:
		halt.As(100, t)
	}
	return err
}

type mach struct {
	ch  chan lem.Message
	ctx lem.Context
	in  chan lem.Message
}

func TypeOf(msg lem.Message) (ret Type) {
	ret = Wrong
	if t, ok := msg["type"].(string); ok {
		ret = TypMap[t]
	}
	return
}

func (m *mach) Do(mod *ir.Module, ld lem.Loader) {
	_run(mod, ld, m.Chan())
}

func (m *mach) Process(msg lem.Message) (ret lem.Message, stop bool) {
	t := TypeOf(msg)
	switch t {
	case Machine:
		if msg["context"] != nil {
			m.ctx = msg["context"].(lem.Context)
		} else {
			stop = true
		}
	case Console:
		fmt.Print(msg["data"])
		if b, _ := msg["ln"].(bool); b {
			fmt.Println()
		}
	case Kernel:
		switch msg["action"].(string) {
		case "load":
			m.ctx.Queue(msg["data"].(string))
		case "event":
			to := msg["timeout"].(int64)
			delay := make(chan int64)
			if to > 0 {
				go func(to int64) {
					<-time.After(time.Millisecond * time.Duration(to))
					delay <- to
				}(to)
			}
			select {
			case nm := <-m.Input():
				ret = make(map[interface{}]interface{})
				ret["type"] = "kernel"
				ret["action"] = "event"
				ret["data"] = nm
			case <-delay:
				ret = make(map[interface{}]interface{})
				ret["type"] = "kernel"
				ret["action"] = "event"
				nm := make(map[interface{}]interface{})
				nm["type"] = "sig"
				nm["sig"] = "none"
				ret["data"] = nm
			}
		case "broadcast":
			m.ctx.Queue(msg["data"])
		default:
			halt.As(100, "unknown method ", msg)
		}
	case Storage:
		os.Mkdir(STORAGE, os.FileMode(0777))
		switch msg["action"].(string) {
		case "store":
			key := msg["key"].(string)
			fn := base64.StdEncoding.EncodeToString([]byte(key))
			if obj := msg["object"]; obj != nil {
				if f, err := os.Create(filepath.Join(STORAGE, fn)); err == nil {
					if data, err := xml.Marshal(obj); err == nil {
						buf := bytes.NewBuffer([]byte(xml.Header))
						buf.Write(data)
						io.Copy(f, buf)
						f.Close()
					} else {
						halt.As(100, err)
					}

				} else {
					halt.As(100, err)
				}
			} else {
				halt.As(100, "nil object")
			}
		case "load":
			key := msg["key"].(string)
			fn := base64.StdEncoding.EncodeToString([]byte(key))
			ret = make(map[interface{}]interface{})
			ret["type"] = "storage"
			ret["key"] = key
			if f, err := os.Open(filepath.Join(STORAGE, fn)); err == nil {
				buf := bytes.NewBuffer(nil)
				io.Copy(buf, f)
				x := &unraw{}
				if err := xml.Unmarshal(buf.Bytes(), x); err == nil {
					ret["object"] = x
				} else {
					halt.As(100, err)
				}
			}
		default:
			halt.As(100, "unknown method ", msg)
		}
	default:
		halt.As(100, "wrong message ", msg)
	}
	return
}

func (m *mach) Input() chan lem.Message {
	if m.in == nil {
		m.in = make(chan lem.Message, 128)
	}
	return m.in
}

func (m *mach) Chan() chan lem.Message {
	if m.ch == nil {
		m.ch = make(chan lem.Message)
		go func(ch chan lem.Message) {
			for {
				msg, stop := m.Process(<-ch)
				if stop {
					break
				}
				ch <- msg
			}
		}(m.ch)
	}
	return m.ch
}

func (m *mach) Stop() {
	msg := make(map[interface{}]interface{})
	msg["type"] = "machine"
	m.Chan() <- msg
}

func init() {
	TypMap = map[string]Type{"console": Console, "machine": Machine, "kernel": Kernel, "storage": Storage}
	lem.Rt = func() lem.Machine {
		return &mach{}
	}
}
