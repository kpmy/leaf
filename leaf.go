package main

import (
	"github.com/kpmy/ypk/halt"
	"leaf/ir"
	"leaf/ir/target"
	_ "leaf/ir/target/yt/z"
	"leaf/lem"
	_ "leaf/lem/ym"
	"leaf/lenin"
	_ "leaf/lenin/rt/dumb"
	_ "leaf/lenin/trav"
	"log"
	"os"
	"path/filepath"
	"unicode"
)

const codeExt = ".li"
const code = ".code"
const SYSTEM = "System"

var root string

func init() {
	root, _ = os.Getwd()
}

func splitName(_n string) (sub []string) {
	n := []rune(_n)
	p := 0
	var ch rune
	var buf []rune
	stop := func() bool {
		return p == len(n)
	}
	up := func() bool {
		ch = n[p]
		p++
		return unicode.IsUpper(ch)
	}
	grow := func() {
		if ch != 0 {
			buf = append(buf, ch)
		}
	}
	low := func() {
		for {
			grow()
			if stop() || up() {
				break
			}
		}
	}
	big := func() {
		for {
			grow()
			if stop() || !up() {
				break
			}
		}
	}
	litOrBig := func() {
		grow()
		if !up() {
			for {
				grow()
				if stop() || up() {
					break
				}
			}
		} else {
			big()
		}
	}
	name := func() string {
		buf = nil
		grow()
		if up() {
			litOrBig()
		} else {
			low()
		}
		return string(buf)
	}
	for p < len(n) {
		sub = append(sub, name())
	}
	return
}
func exists(fullpath string) (ret bool) {
	ret = false
	if fi, err := os.Stat(fullpath); err == nil {
		ret = !fi.IsDir()
	}
	return
}

func load(name string) (ret *ir.Module, err error) {

	if t, err := os.Open(name + ".li"); err == nil {
		defer t.Close()
		ret = target.Old(t)
	}
	return
}

func do(fullpath string) {
	if li, err := os.Open(fullpath); err == nil {
		m := target.Old(li)
		mach := lem.Run()
		lenin.Do(m, load, mach.Chan())
	} else {
		log.Fatal(err)
	}
}

func open(do func(string), name string, path ...string) {
	if path != nil {
		try := filepath.Join(root, filepath.Join(path...), code, name+codeExt)
		if exists(try) {
			do(try)
		}
	} else { //System?
		try := filepath.Join(root, code, name+codeExt)
		if exists(try) {
			do(try)
		} else {
			try := filepath.Join(root, SYSTEM, code, name+codeExt)
			if exists(try) {
				do(try)
			}
		}
	}
}

func find(name string, do func(string)) {
	n := splitName(name)
	if len(n) > 0 {
		mod := n[len(n)-1]
		var sub []string
		for i := len(n) - 2; i >= 0; i-- {
			sub = append(sub, n[i])
		}
		if len(sub) == 1 && sub[0] == SYSTEM {
			sub = nil
		}
		open(do, mod, sub...)
	} else {
		halt.As(100, "wrong name")
	}
}

func main() {
	log.Println("Leaf framework, pk, 20150703")
	find("Init", do)
}
