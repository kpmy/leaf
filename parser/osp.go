package parser

import (
	"io"
	"leaf/scanner"
	"log"
)

type Parser interface {
	Compile(io.Reader)
}

type pr struct {
	oss scanner.Scanner
	sym scanner.Symbol
}

func (p *pr) get() {
	p.sym = p.oss.Get()
}

func (p *pr) check(s scanner.Symbol, msg string) {
	if s == p.sym {
		p.get()
	} else {
		p.oss.Mark(msg)
	}
}

func (p *pr) module() {
	var modid string
	if p.sym == scanner.Module {
		p.get()
		if p.sym == scanner.Ident {
			p.get()
			log.Println("Compiling", p.oss.Id()+"...")
			modid = p.oss.Id()
		} else {
			p.oss.Mark("ident?")
		}
		p.check(scanner.Semicolon, "; expected")
		if p.sym == scanner.Begin {
			p.get()
		}
		p.check(scanner.End, "no END")
		if p.sym == scanner.Ident {
			if modid != p.oss.Id() {
				p.oss.Mark("no match")
			}
			p.get()
		} else {
			p.oss.Mark("ident?")
		}
		if p.sym != scanner.Period {
			p.oss.Mark(". ?")
		}
		if p.oss.Error() == nil {
			log.Println("Compiled")
		}
	} else {
		p.oss.Mark("MOD?")
	}
}

func (p *pr) Compile(rd io.Reader) {
	p.oss.Init(rd)
	p.get()
	p.module()
}

func New() Parser {
	return &pr{oss: scanner.New()}
}
