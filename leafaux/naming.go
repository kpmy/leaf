package leafaux

import (
	"log"
	"unicode"
)

func SplitName(_n string) (sub []string) {
	n := []rune(_n)
	p := 0
	var ch rune
	var buf []rune
	stop := func() bool {
		return p == len(n)
	}
	up := func() bool {
		ch = n[p]
		if ch == '.' {
			log.Fatal("invalid character `.`")
		}
		p++
		return unicode.IsUpper(ch)
	}
	grow := func() {
		if ch != 0 {
			buf = append(buf, ch)
		}
	}
	low := func() {
		for {
			grow()
			if stop() || up() {
				break
			}
		}
	}
	big := func() {
		for {
			grow()
			if stop() || !up() {
				break
			}
		}
	}
	litOrBig := func() {
		grow()
		if !up() {
			for {
				grow()
				if stop() || up() {
					break
				}
			}
		} else {
			big()
		}
	}
	name := func() string {
		buf = nil
		grow()
		if up() {
			litOrBig()
		} else {
			low()
		}
		return string(buf)
	}
	for p < len(n) {
		sub = append(sub, name())
	}
	return
}
