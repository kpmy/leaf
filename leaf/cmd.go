package main

import (
	"bufio"
	"flag"
	"github.com/kpmy/ypk/assert"
	"leaf/li"
	_ "leaf/li/trav"
	"leaf/parser"
	"leaf/scanner"
	"leaf/target"
	_ "leaf/target/yt"
	"log"
	"os"
)

var name string

func init() {
	flag.StringVar(&name, "i", "Simple0", "-i name.ext")
}

func main() {
	log.Println("Leaf compiler, pk, 20150529")
	flag.Parse()
	assert.For(name != "", 20)
	log.Println(name)
	sname := name + ".lf"
	if f, err := os.Open(sname); err == nil {
		defer f.Close()
		p := parser.ConnectTo(scanner.ConnectTo(bufio.NewReader(f)))
		ir, _ := p.Module()
		if t, err := os.Create(name + ".li"); err == nil {
			target.New(ir, t)
			t.Close()
		}
		if z, err := os.Open(name + ".li"); err == nil {
			ir := target.Old(z)
			li.Do(ir)
		}
	} else {
		log.Fatal(err)
	}
}
