package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/kpmy/ypk/assert"
	"io"
	"leaf/parser"
	"leaf/scanner"
	"log"
	"os"
)

var name string

func init() {
	flag.StringVar(&name, "i", "test0.lf", "-i name.ext")
}

func trivia(rd io.RuneReader) {
	sc := scanner.ConnectTo(rd)
	buf := make([]scanner.Sym, 0)
	for sc.Error() == nil {
		buf = append(buf, sc.Get())
	}
	fmt.Println("SCANNER OUTPUT")
	for _, v := range buf {
		switch v.Code {
		case scanner.Ident:
			fmt.Print(`@` + v.String)
		case scanner.Delimiter:
			fmt.Println()
		case scanner.Separator:
			fmt.Print(" ")
		case scanner.String:
			fmt.Print(`"` + v.String + `"`)
		case scanner.Number:
			fmt.Print(v.String)
		default:
			fmt.Print(v.Code)
		}
	}
}

func main() {
	log.Println("Leaf compiler, pk, 20150529")
	flag.Parse()
	assert.For(name != "", 20)
	log.Println(name)
	if f, err := os.Open(name); err == nil {
		defer f.Close()
		trivia(bufio.NewReader(f))
	} else {
		log.Fatal(err)
	}
	if f, err := os.Open(name); err == nil {
		defer f.Close()
		sc := scanner.ConnectTo(bufio.NewReader(f))
		p := parser.ConnectTo(sc)
		p.Module()
	} else {
		log.Fatal(err)
	}

}
