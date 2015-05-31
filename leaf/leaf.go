package main

import (
	"bufio"
	"flag"
	"github.com/kpmy/ypk/assert"
	"leaf/parser"
	"log"
	"os"
)

var name string

func init() {
	flag.StringVar(&name, "i", "test0.leaf", "-i name.ext")
}

func main() {
	log.Println("Leaf compiler, pk, 20150529")
	flag.Parse()
	assert.For(name != "", 20)
	log.Println(name)
	if f, err := os.Open(name); err == nil {
		defer f.Close()
		p := parser.New()
		p.Compile(bufio.NewReader(f))
	} else {
		log.Fatal(err)
	}
}
