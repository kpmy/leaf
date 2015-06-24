package leaf

import (
	"bufio"
	"fmt"
	"leaf/ir"
	"leaf/lenin"
	_ "leaf/lenin/trav"
	"leaf/parser"
	"leaf/scanner"
	"leaf/target"
	_ "leaf/target/yt/z"
	"log"
	"os"
	"strconv"
	"testing"
)

func TestScanner(t *testing.T) {
	if f, err := os.Open("test-scanner.lf"); err == nil {
		defer f.Close()
		sc := scanner.ConnectTo(bufio.NewReader(f))
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
		sname := mname + ".lf"
		if _, err = os.Stat(sname); err == nil {
			if f, err := os.Open(sname); err == nil {
				defer f.Close()
				p := parser.ConnectTo(scanner.ConnectTo(bufio.NewReader(f)))
				ir, _ := p.Module()
				if t, err := os.Create(mname + ".li"); err == nil {
					defer t.Close()
					target.New(ir, t)
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
				ir := target.Old(f)
				if t, err := os.Create(mname + ".lio"); err == nil {
					defer t.Close()
					target.New(ir, t)
				}
			}
		}
	}
}

func TestInterp(t *testing.T) {
	var err error
	for i := int64(0); err == nil; i++ {
		mname := "Test" + strconv.FormatInt(i, 16)
		sname := mname + ".lf"
		if _, err = os.Stat(sname); err == nil {
			if f, err := os.Open(sname); err == nil {
				defer f.Close()
				p := parser.ConnectTo(scanner.ConnectTo(bufio.NewReader(f)))
				ir, _ := p.Module()
				lenin.Do(ir)
			}
		}
	}
}

func TestCollection(t *testing.T) {
	var err error
	for i := int64(0); err == nil; i++ {
		mname := "Test" + strconv.FormatInt(i, 16)
		sname := mname + ".lc"
		if _, err = os.Stat(sname); err == nil {
			if f, err := os.Open(sname); err == nil {
				defer f.Close()
				rd := bufio.NewReader(f)
				for err == nil {
					p := parser.ConnectTo(scanner.ConnectTo(rd))
					var ast *ir.Module
					if ast, err = p.Module(); err == nil {
						if t, err := os.Create(ast.Name + ".li"); err == nil {
							target.New(ast, t)
							t.Close()
						}
						if t, err := os.Open(ast.Name + ".li"); err == nil {
							defer t.Close()
							ast := target.Old(t)
							lenin.Do(ast)
						}
					}
				}
			}
		}
	}
}
