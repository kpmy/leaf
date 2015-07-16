package main

import (
	"flag"
	"leaf/leaf"
	"log"
	"strings"
)

var build string
var debug bool

func init() {
	flag.StringVar(&build, "b", "", "-b=Module")
	flag.BoolVar(&debug, "debug", false, "-debug=true/false")
}

func main() {
	log.Println("Leaf framework, pk, 20150703")
	flag.Parse()
	//build = "TestEvents"
	//debug = true
	switch {
	case build != "":
		log.Println("build", build)
		for _, n := range strings.Split(build, " ") {
			leaf.DoBuild(n, debug)
		}
	default:
		leaf.Do("Init", debug)
	}
}
