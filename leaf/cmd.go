package main

import (
	"bufio"
	"flag"
	"github.com/kpmy/ypk/assert"
	"leaf/parser"
	"leaf/scanner"
	"log"
	"os"
)

var name string

func init() {
	flag.StringVar(&name, "i", "simple.lf", "-i name.ext")
}

func main() {
	log.Println("Leaf compiler, pk, 20150529")
	flag.Parse()
	assert.For(name != "", 20)
	log.Println(name)
	if f, err := os.Open(name); err == nil {
		defer f.Close()
		sc := scanner.ConnectTo(bufio.NewReader(f))
		p := parser.ConnectTo(sc)
		p.Module()
	} else {
		log.Fatal(err)
	}
}
