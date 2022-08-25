package sortedIpHeap

import (
	"container/heap"
	"net/netip"
)

type sortedIpHeap []netip.Addr

func New() *sortedIpHeap {
	h := &sortedIpHeap{}

	heap.Init(h)

	return h
}

func (h sortedIpHeap) Len() int {
	return len(h)
}

func (h sortedIpHeap) Less(i, j int) bool {
	return h[i].Less(h[j])
}

func (h sortedIpHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
}

func (h *sortedIpHeap) Push(x interface{}) {
	*h = append(*h, x.(netip.Addr))
}

func (h *sortedIpHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}
