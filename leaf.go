package main

import (
	"flag"
	"github.com/kpmy/leaf/leafaux"
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
			leafaux.DoBuild(n, debug)
		}
	default:
		leafaux.Do("Init", debug)
	}
}
