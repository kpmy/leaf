package demo

import (
	"bufio"
	"fmt"
	"leaf/parser"
	"leaf/scanner"
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
	for i := 0; err == nil; i++ {
		name := "test" + "-" + strconv.Itoa(i) + ".lf"
		if _, err = os.Stat(name); err == nil {
			if f, err := os.Open(name); err == nil {
				defer f.Close()
				p := parser.ConnectTo(scanner.ConnectTo(bufio.NewReader(f)))
				p.Module()
			}
		}
	}
}
