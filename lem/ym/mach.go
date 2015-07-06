package ym

import (
	"fmt"
	"github.com/kpmy/ypk/halt"
	"leaf/lem"
	"leaf/lenin/rt"
)

type Type int

const (
	Wrong Type = iota
	Machine
	Console
	Kernel
)

var TypMap map[string]Type

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
	switch TypeOf(msg) {
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
	default:
		halt.As(100, "wrong message ")
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
	TypMap = map[string]Type{"console": Console, "machine": Machine, "kernel": Kernel}
	lem.Rt = func() lem.Machine {
		return &mach{}
	}
}
