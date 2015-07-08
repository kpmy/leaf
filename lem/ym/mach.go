package ym

import (
	"bytes"
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"github.com/kpmy/ypk/halt"
	"io"
	"leaf/lem"
	"leaf/lenin/rt"
	"os"
	"path/filepath"
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

type raw struct {
	rd io.Reader
}

func (r *raw) Convert() {}

type mach struct {
	ch  chan rt.Message
	ctx rt.Context
}

func TypeOf(msg rt.Message) (ret Type) {
	ret = Wrong
	if t, ok := msg["type"].(string); ok {
		ret = TypMap[t]
	}
	return
}

func (m *mach) Do(msg rt.Message) (ret rt.Message, stop bool) {
	t := TypeOf(msg)
	switch t {
	case Machine:
		if msg["context"] != nil {
			m.ctx = msg["context"].(rt.Context)
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
		}
	case Storage:
		os.Mkdir(STORAGE, os.FileMode(0777))
		switch msg["action"].(string) {
		case "store":
			key := msg["key"].(string)
			fn := base64.StdEncoding.EncodeToString([]byte(key))
			if obj := msg["object"]; obj != nil {
				obj.(lem.Object).Convert()
				if f, err := os.Create(filepath.Join(STORAGE, fn)); err == nil {
					data, _ := xml.Marshal(obj)
					buf := bytes.NewBuffer([]byte(xml.Header))
					buf.Write(data)
					io.Copy(f, buf)
					f.Close()
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
			ret["key"] = key
			if f, err := os.Open(filepath.Join(STORAGE, fn)); err == nil {
				ret["object"] = &raw{rd: f}
			}
		}
	default:
		halt.As(100, "wrong message ", msg)
	}
	return
}

func (m *mach) Chan() chan rt.Message {
	if m.ch == nil {
		m.ch = make(chan rt.Message)
		go func(ch chan rt.Message) {
			for {
				msg, stop := m.Do(<-ch)
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
