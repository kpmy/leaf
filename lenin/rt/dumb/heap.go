package dumb

import (
	"fmt"
	"leaf/lenin/trav"
)

type heap struct {
	data map[int64]*trav.Any
	next int64
}

func (h *heap) New() (ret int64) {
	h.data[h.next] = &trav.Any{}
	ret = h.next
	h.next++
	return
}

type heapy struct {
	adr int64
	h   *heap
}

func (h *heapy) Get() *trav.Any {
	return h.h.data[h.adr]
}

func (h *heapy) Set(x *trav.Any) {
	fmt.Println("heap touch", h.adr, x)
	h.h.data[h.adr] = x
}

func newHeap() *heap {
	ret := &heap{}
	ret.data = make(map[int64]*trav.Any)
	ret.next = 4096
	return ret
}
