package leaf

import (
	"bufio"
	"fmt"
	"leaf/ir"
	code "leaf/ir/target"
	_ "leaf/ir/target/yt/z"
	"leaf/lead"
	def "leaf/lead/target"
	_ "leaf/lead/target/tt"
	"leaf/leap"
	"leaf/lem"
	_ "leaf/lem/ym"
	"leaf/lenin"
	_ "leaf/lenin/rt/dumb"
	_ "leaf/lenin/trav"
	scanner "leaf/lss"
	"log"
	"os"
	"strconv"
	"testing"
)

func resolve(name string) (ret *ir.Import, err error) {
	if d, err := os.Open(name + ".ld"); err == nil {
		p := lead.ConnectTo(scanner.ConnectTo(bufio.NewReader(d)), resolve)
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
				p := leap.ConnectTo(scanner.ConnectTo(bufio.NewReader(f)), resolve)
				ir, _ := p.Module()
				if t, err := os.Create(mname + ".li"); err == nil {
					defer t.Close()
					code.New(ir, t)
				}
				if t, err := os.Create(mname + ".ld"); err == nil {
					defer t.Close()
					def.New(ir, t)
				}
				if t, err := os.Create(mname + ".lm"); err == nil {
					defer t.Close()
					lem.New(ir, t)
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
				p := leap.ConnectTo(scanner.ConnectTo(bufio.NewReader(f)), resolve)
				ir, _ := p.Module()
				mach := lem.Run()
				lenin.Debug = true
				lenin.Do(ir, load, mach.Chan())
			}
		}
	}
}

func TestCollection(t *testing.T) {
	lenin.Debug = true
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
					p := leap.ConnectTo(scanner.ConnectTo(rd), resolve)
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
							p := lead.ConnectTo(scanner.ConnectTo(bufio.NewReader(d)), resolve)
							p.Import()
						}
						if t, err := os.Create(ast.Name + ".lm"); err == nil {
							defer t.Close()
							lem.New(ast, t)
						}
						if t, err := os.Open(ast.Name + ".li"); err == nil {
							defer t.Close()
							ast := code.Old(t)
							mach := lem.Run()
							lenin.Do(ast, load, mach.Chan())
						}
					}
				}
			}
		}
	}
}
