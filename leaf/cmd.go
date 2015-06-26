package main

import (
	"bufio"
	"flag"
	"github.com/kpmy/ypk/assert"
	code "leaf/ir/target"
	_ "leaf/ir/target/yt/z"
	def "leaf/lead/target"
	_ "leaf/lead/target/tt"
	"leaf/leap"
	"leaf/lenin"
	_ "leaf/lenin/trav"
	scanner "leaf/lss"
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
	sname := name + ".lf"
	log.Println(name, "running...")
	if f, err := os.Open(sname); err == nil {
		defer f.Close()
		p := leap.ConnectTo(scanner.ConnectTo(bufio.NewReader(f)))
		ir, _ := p.Module()
		if d, err := os.Open(name + ".ld"); err == nil {
			d.Close()
		}
		if d, err := os.Create(name + ".ld"); err == nil {
			def.New(ir, d)
			d.Close()
		}
		if t, err := os.Create(name + ".li"); err == nil {
			code.New(ir, t)
			t.Close()
		}
		if z, err := os.Open(name + ".li"); err == nil {
			ir := code.Old(z)
			lenin.Do(ir)
		}
		log.Println(name, "end.")
	} else {
		log.Fatal(err)
	}
}
