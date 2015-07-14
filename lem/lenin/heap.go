package lenin

import (
	"fmt"
	"leaf/lem"
)

type heap struct {
	data map[int64]*Any
	next int64
}

func (h *heap) New() (ret int64) {
	h.data[h.next] = &Any{}
	ret = h.next
	h.next++
	return
}

type heapy struct {
	adr int64
	h   *heap
}

func (h *heapy) Get() *Any {
	return h.h.data[h.adr]
}

func (h *heapy) Set(x *Any) {
	if lem.Debug {
		fmt.Println("heap touch", fmt.Sprintf("%X", h.adr), x)
	}
	h.h.data[h.adr] = x
	if lem.Debug {
		fmt.Println(h.h)
	}
}

func newHeap() *heap {
	ret := &heap{}
	ret.data = make(map[int64]*Any)
	ret.next = 4096
	return ret
}
