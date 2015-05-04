package flowgraph

import (
	"sort"
	"sync"
	"sync/atomic"
)

var poolQsortSz int64
var poolQsortMu = &sync.Mutex{}

type Interface2 interface {
	// sort.Interface is borrowed from the sort package.
	sort.Interface
	// Sorted tests if slice is sorted.
	Sorted() bool
	// Sorted tests if original array is sorted.
	OrigSorted() bool
	// SubSlice returns a sub-slice.
	SubSlice(n, m int) Datum
	// Orig returns original slice
	Orig() []int
	// Depth returns the depth of a recursive sort
	Depth() int64
	// ID returns a unique ID for the object
	ID() int64
}

func qsortFire (n *Node) {

	a := n.Srcs[0]
	steerAck := a.Ack2 != nil
	a.SendAck(n) // write early to let flow go on

	x := n.Dsts[0]
	a.NoOut = true
	if _,ok := a.Val.(Interface2); !ok {
		n.Errorf("not of type Interface2 (%T)\n", a.Val)
		return
	}

	d := a.Val.(Interface2)
	l := d.Len()

	var poolSz int64
	freeNode := func(num int) bool {
		poolQsortMu.Lock()
		defer poolQsortMu.Unlock()
		if poolQsortSz>1 {
			poolQsortSz -= int64(num)
			n.Tracef("\tpool %s\n", func() string {var s string; for i:=int64(0); i<poolQsortSz; i++ { s += "*" }; return s}())
			poolSz = poolQsortSz+int64(num)
			return true
		}
		poolSz = poolQsortSz
		return false
	}

	snap := func() {
		n.Tracef("Original(%p) sorted %t, Sliced sorted %t, poolsz=%d, depth=%d, id=%d, len=%d\n", d.Orig(), d.OrigSorted(), d.Sorted(), poolSz, d.Depth(), d.ID(), d.Len())
	}

	if l <= 4096 || !freeNode(2) {
		snap()
		n.Tracef("Original(%p) call sort.Sort\n", d.Orig())
		sort.Sort(d)
		x.Val=x.AckWrap(d)
		x.SendData(n)
		x.NoOut = true
		if steerAck {
			atomic.AddInt64(&poolQsortSz, 1)
			n.Tracef("\tpool %s\n", func() string {var s string; for i:=int64(0); i<poolQsortSz; i++ { s += "*" }; return s}())
		}
		return
	}

	snap()
	mlo,mhi := doPivot(d, 0, l)
	c := 0
	xData := x.Data
	xName := x.Name
	x.Data = a.Data // recurse
	x.Name = x.Name+"("+a.Name+")"
	if mlo>0 {
		n.Tracef("Original(%p) recurse left [0:%d]\n", d.Orig(), mlo)
		x.Val = x.AckWrap(d.SubSlice(0, mlo))
		x.SendData(n)
		c++
	}
	if l-mhi>0 {
		n.Tracef("Original(%p) recurse right [%d:%d]\n", d.Orig(), mhi, l)
		x.Val = x.AckWrap(d.SubSlice(mhi, l))
		x.SendData(n)
		c++
	}
	x.Data = xData
	x.Name = xName
	
	if steerAck {
		atomic.AddInt64(&poolQsortSz, 1)
		n.Tracef("\tpool %s\n", func() string {var s string; for i:=int64(0); i<poolQsortSz; i++ { s += "*" }; return s}())
	}
	x.RdyCnt = c
	x.NoOut = true

}

// FuncQsort recursively implements a quicksort with goroutines (x=qsort(a)).
func FuncQsort(a, x Edge, poolSz int) []Node {
	
	// Make a pool of qsort nodes that can be dynamically used, 
	// and reserve one for the front end input into this dynamically 
	// extruded flowgraph.
	n := MakeNodes(poolSz)

	poolQsortSz = int64(poolSz)-1
	for i:=0; i<poolSz; i++ {
		aa, xx := a,x  // make a copy of the Edge's for each one
		n[i] = MakeNodePool("qsort", []*Edge{&aa}, []*Edge{&xx}, nil, qsortFire)
	}
	return n

}
