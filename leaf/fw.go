package leaf

import (
	"bufio"
	"errors"
	"fmt"
	"leaf/ir"
	"leaf/ir/target"
	_ "leaf/ir/target/yt/z"
	"leaf/leac/def"
	_ "leaf/leac/def/tt"
	"leaf/leac/leap"
	"leaf/leac/lss"
	"leaf/lem"
	_ "leaf/lem/lenin"
	"log"
	"os"
	"path/filepath"
)

const codeExt = ".li"
const sourceExt = ".leaf"
const defExt = ".ld"
const CODE = ".code"
const SYSTEM = "System"
const SOURCE = ""
const DEF = ".def"

var debug bool
var root string

func init() {
	root, _ = os.Getwd()

}
func exists(fullpath string) (ret bool) {
	ret = false
	if fi, err := os.Stat(fullpath); err == nil {
		ret = !fi.IsDir()
	}
	return
}

func load(n string) (ret *ir.Module, rerr error) {
	doFind(n, func(fullpath string) {
		if t, err := os.Open(fullpath); err == nil {
			defer t.Close()
			ret = target.Old(t)
		} else {
			rerr = err
		}
	})
	if ret == nil && rerr == nil {
		rerr = errors.New("file not found")
	}
	return
}

func do(fullpath string) {
	lem.Debug = debug
	if li, err := os.Open(fullpath); err == nil {
		m := target.Old(li)
		mach := lem.Start()
		mach.Do(m, load)
		mach.Stop()
	} else {
		log.Fatal(err)
	}
}

func create(do func(string), typ string, name string, ext string, path ...string) {
	if path != nil {
		cat := filepath.Join(root, filepath.Join(path...), typ)
		if err := os.MkdirAll(cat, os.FileMode(0777)); err == nil {
			try := filepath.Join(root, filepath.Join(path...), typ, name+ext)
			do(try)
		}
	} else { //System?
		cat := filepath.Join(root, typ)
		if err := os.MkdirAll(cat, os.FileMode(0777)); err == nil {
			try := filepath.Join(root, typ, name+ext)
			do(try)
		}
	}
}

func open(do func(string), typ string, name string, ext string, path ...string) {
	if path != nil {
		try := filepath.Join(root, filepath.Join(path...), typ, name+ext)
		if exists(try) {
			do(try)
		}
	} else { //System?
		try := filepath.Join(root, typ, name+ext)
		if exists(try) {
			do(try)
		} else {
			try := filepath.Join(root, SYSTEM, typ, name+ext)
			if exists(try) {
				do(try)
			}
		}
	}
}

func doFind(name string, do func(string)) {
	n := SplitName(name)
	if len(n) > 0 {
		mod := n[len(n)-1]
		var sub []string
		for i := len(n) - 2; i >= 0; i-- {
			sub = append(sub, n[i])
		}
		if len(sub) == 1 && sub[0] == SYSTEM {
			sub = nil
		}
		open(do, CODE, mod, codeExt, sub...)
	} else {
		log.Fatal("wrong name")
	}
}
func Do(name string, _debug bool) {
	doFind(name, do)
}

func DoBuild(name string, _debug bool) {
	var mod string
	var sub []string
	var resolve func(name string) (ret *ir.Import, err error)
	resolve = func(name string) (ret *ir.Import, err error) {
		var mod string
		var sub []string
		n := SplitName(name)
		if len(n) > 0 {
			mod = n[len(n)-1]
			for i := len(n) - 2; i >= 0; i-- {
				sub = append(sub, n[i])
			}
			if len(sub) == 1 && sub[0] == SYSTEM {
				sub = nil
			}
			open(func(fullpath string) {
				if d, err := os.Open(fullpath); err == nil {
					p := leap.ConnectToDef(lss.ConnectTo(bufio.NewReader(d)), resolve)
					ret, _ = p.Import()
				}
			}, DEF, mod, defExt, sub...)
		} else {
			log.Fatal("wrong name")
		}
		return
	}
	var compile func(string)
	compile = func(fullpath string) {
		var msg string
		if f, err := os.Open(fullpath); err == nil {
			defer f.Close()
			log.Println("compiling", name)
			p := leap.ConnectToMod(lss.ConnectTo(bufio.NewReader(f)), resolve)
			code, _ := p.Module()
			msg = fmt.Sprint("compiled ", name)
			create(func(fullpath string) {
				if f, err := os.Create(fullpath); err == nil {
					defer f.Close()
					target.New(code, f)
					msg = fmt.Sprint(msg, " code")
				} else {
					log.Fatal(err)
				}
			}, CODE, mod, codeExt, sub...)

			create(func(fullpath string) {
				if f, err := os.Create(fullpath); err == nil {
					defer f.Close()
					def.New(code, f)
					msg = fmt.Sprint(msg, " def")
				} else {
					log.Fatal(err)
				}
			}, DEF, mod, defExt, sub...)
			log.Println(msg, "ok")
		} else {
			log.Fatal(err)
		}
	}
	n := SplitName(name)
	if len(n) > 0 {
		mod = n[len(n)-1]
		for i := len(n) - 2; i >= 0; i-- {
			sub = append(sub, n[i])
		}
		if len(sub) == 1 && sub[0] == SYSTEM {
			sub = nil
		}
		open(compile, SOURCE, mod, sourceExt, sub...)
	} else {
		log.Fatal("wrong name")
	}
}
