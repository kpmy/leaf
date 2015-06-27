package main

import (
	"bufio"
	"flag"
	"github.com/kpmy/ypk/assert"
	"leaf/ir"
	code "leaf/ir/target"
	_ "leaf/ir/target/yt/z"
	"leaf/lead"
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

func main() {
	log.Println("Leaf compiler, pk, 20150529")
	flag.Parse()
	assert.For(name != "", 20)
	sname := name + ".lf"
	log.Println(name, "running...")
	if f, err := os.Open(sname); err == nil {
		defer f.Close()
		p := leap.ConnectTo(scanner.ConnectTo(bufio.NewReader(f)), resolve)
		ir, _ := p.Module()
		if d, err := os.Open(name + ".ld"); err == nil {
			log.Println("definition already exists")
			d.Close()
		}
		if t, err := os.Create(name + ".li"); err == nil {
			code.New(ir, t)
			t.Close()
		}
		if z, err := os.Open(name + ".li"); err == nil {
			ir := code.Old(z)
			if d, err := os.Create(name + ".ld"); err == nil {
				def.New(ir, d)
				d.Close()
			}
			lenin.Do(ir, load)
		}
		log.Println(name, "end.")
	} else {
		log.Fatal(err)
	}
}
