package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/kpmy/ypk/assert"
	"leaf/ir"
	code "leaf/ir/target"
	_ "leaf/ir/target/yt/z"
	"leaf/lead"
	def "leaf/lead/target"
	_ "leaf/lead/target/tt"
	"leaf/leap"
	"leaf/lem"
	_ "leaf/lem/lenin"
	scanner "leaf/lss"
	"log"
	"os"
)

var name string
var debug bool
var interp bool

func init() {
	flag.StringVar(&name, "source", "Simple", "-source=name")
	flag.BoolVar(&debug, "debug", false, "-debug=true/false")
	flag.BoolVar(&interp, "int", false, "-int=true/false")
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
	sname := name + ".leaf"
	log.Println(name, "compiling...")
	if f, err := os.Open(sname); err == nil {
		defer f.Close()
		p := leap.ConnectTo(scanner.ConnectTo(bufio.NewReader(f)), resolve)
		if ir, err := p.Module(); err == nil {
			defer func() {
				if x := recover(); x != nil {
					fmt.Println(x)
				}
			}()
			if d, err := os.Open(name + ".ld"); err == nil {
				log.Println("definition already exists")
				d.Close()
			}
			if t, err := os.Create(name + ".li"); err == nil {
				code.New(ir, t)
				t.Close()
			}
		} else {
			panic(err)
		}
		if z, err := os.Open(name + ".li"); err == nil {
			ir := code.Old(z)
			if d, err := os.Create(name + ".ld"); err == nil {
				def.New(ir, d)
				d.Close()
			}
			if interp {
				log.Println(name, "running...")
				mach := lem.Start()
				lem.Debug = debug
				mach.Do(ir, load)
				mach.Stop()
			}
		}
	} else {
		log.Fatal(err)
	}
}
