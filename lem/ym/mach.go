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
	Console
)

var TypMap map[string]Type

type mach struct {
}

func TypeOf(msg rt.Message) (ret Type) {
	ret = Wrong
	if t, ok := msg["type"].(string); ok {
		ret = TypMap[t]
	}
	return
}

func (m *mach) Do(msg rt.Message) (ret rt.Message) {
	switch TypeOf(msg) {
	case Console:
		fmt.Print(msg["data"])
		if b, _ := msg["ln"].(bool); b {
			fmt.Println()
		}
	default:
		halt.As(100, "wrong message ")
	}
	return
}

func (m *mach) Chan() (ch chan rt.Message) {
	ch = make(chan rt.Message)
	go func(ch chan rt.Message) {
		for {
			ch <- m.Do(<-ch)
		}
	}(ch)
	return
}

func init() {
	TypMap = map[string]Type{"console": Console}
	lem.Rt = func() lem.Machine {
		return &mach{}
	}
}
