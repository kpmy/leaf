package main

import (
	"bufio"
	"fmt"
	"github.com/kpmy/leaf/ir"
	code "github.com/kpmy/leaf/ir/target"
	_ "github.com/kpmy/leaf/ir/target/yt/z"
	"github.com/kpmy/leaf/leac/def"
	_ "github.com/kpmy/leaf/leac/def/tt"
	"github.com/kpmy/leaf/leac/leap"
	scanner "github.com/kpmy/leaf/leac/lss"
	"github.com/kpmy/leaf/lem"
	_ "github.com/kpmy/leaf/lem/lenin"
	"log"
	"os"
	"strconv"
	"testing"
)

func resolve(name string) (ret *ir.Import, err error) {
	if d, err := os.Open(name + ".ld"); err == nil {
		p := leap.ConnectToDef(scanner.ConnectTo(bufio.NewReader(d)), resolve)
		ret, _ = p.Import()
	}
	return
}

func load(name string) (ret *ir.Module, err error) {
	if t, err := os.Open(name + ".li"); err == nil {
		defer t.Close()
		ret = code.Old(t)
	}
	return
}

func TestScanner(t *testing.T) {
	if f, err := os.Open("test-scanner.leaf"); err == nil {
		defer f.Close()
		sc := scanner.ConnectTo(bufio.NewReader(f))
		sc.Init(scanner.Module)
		buf := make([]scanner.Sym, 0)
		for sc.Error() == nil {
			buf = append(buf, sc.Get())
		}
		fmt.Println("SCANNER OUTPUT")
		for _, v := range buf {
			switch v.Code {
			case scanner.Ident:
				fmt.Print(`@` + v.Str)
			case scanner.Delimiter:
				fmt.Println()
			case scanner.Separator:
				fmt.Print(" ")
			case scanner.String:
				fmt.Print(`"` + v.Str + `"`)
			case scanner.Number:
				fmt.Print(v.Str)
			default:
				fmt.Print(v.Code)
			}
		}
	} else {
		log.Fatal(err)
	}
}

func TestParser(t *testing.T) {
	var err error
	for i := int64(0); err == nil; i++ {
		mname := "Test" + strconv.FormatInt(i, 16)
		sname := mname + ".leaf"
		if _, err = os.Stat(sname); err == nil {
			if f, err := os.Open(sname); err == nil {
				defer f.Close()
				p := leap.ConnectToMod(scanner.ConnectTo(bufio.NewReader(f)), resolve)
				ir, _ := p.Module()
				if t, err := os.Create(mname + ".li"); err == nil {
					defer t.Close()
					code.New(ir, t)
				}
				if t, err := os.Create(mname + ".ld"); err == nil {
					defer t.Close()
					def.New(ir, t)
				}
			}
		}
	}
}

func TestAST(t *testing.T) {
	var err error
	for i := int64(0); err == nil; i++ {
		mname := "Test" + strconv.FormatInt(i, 16)
		sname := mname + ".li"
		if _, err = os.Stat(sname); err == nil {
			if f, err := os.Open(sname); err == nil {
				defer f.Close()
				ir := code.Old(f)
				if t, err := os.Create(mname + ".lio"); err == nil {
					defer t.Close()
					code.New(ir, t)
				}
			}
		}
	}
}

func TestInterp(t *testing.T) {
	var err error
	for i := int64(0); err == nil; i++ {
		mname := "Test" + strconv.FormatInt(i, 16)
		sname := mname + ".leaf"
		if _, err = os.Stat(sname); err == nil {
			if f, err := os.Open(sname); err == nil {
				defer f.Close()
				p := leap.ConnectToMod(scanner.ConnectTo(bufio.NewReader(f)), resolve)
				ir, _ := p.Module()
				mach := lem.Start()
				lem.Debug = true
				mach.Do(ir, load)
			}
		}
	}
}

func TestCollection(t *testing.T) {
	lem.Debug = true
	var err error
	for i := int64(0); err == nil; i++ {
		mname := "Test" + strconv.FormatInt(i, 16)
		sname := mname + ".leac"
		if _, err = os.Stat(sname); err == nil {
			if f, err := os.Open(sname); err == nil {
				defer f.Close()
				rd := bufio.NewReader(f)
				for err == nil {
					fmt.Println()
					p := leap.ConnectToMod(scanner.ConnectTo(rd), resolve)
					var ast *ir.Module
					if ast, err = p.Module(); err == nil {
						if t, err := os.Create(ast.Name + ".li"); err == nil {
							code.New(ast, t)
							t.Close()
						}
						if t, err := os.Create(ast.Name + ".ld"); err == nil {
							def.New(ast, t)
							t.Close()
						}
						if d, err := os.Open(ast.Name + ".ld"); err == nil {
							p := leap.ConnectToDef(scanner.ConnectTo(bufio.NewReader(d)), resolve)
							p.Import()
						}
						if t, err := os.Open(ast.Name + ".li"); err == nil {
							defer t.Close()
							ast := code.Old(t)
							mach := lem.Start()
							mach.Do(ast, load)
							mach.Stop()
						}
					}
				}
			}
		}
	}
}
