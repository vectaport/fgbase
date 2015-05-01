package flowgraph

import (
	"sync"
	"sync/atomic"
)

var poolQsortSz int64
var poolQsortMu = &sync.Mutex{}

type Interface2 interface {
	// Interface is borrowed from the sort package.
	Interface
	// Sorted tests if array is sorted.
	Sorted() bool
	// SubSlice returns a sub-slice.
	SubSlice(n, m int) Datum
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

	l := Len(a.Val)
	d := a.Val.(Interface2)

	var freeNode = func(num int) bool {
		poolQsortMu.Lock()
		defer poolQsortMu.Unlock()
		if poolQsortSz>1 {
			poolQsortSz -= int64(num)
			n.Tracef("poolQsortSz-=%d to %d\n", num, poolQsortSz)		
			return true
		}
		return false
	}

	if l <= 1024 || !freeNode(2) {
		Sort(d)
		x.Val=x.AckWrap(d)
		x.SendData(n)
		x.NoOut = true
		if steerAck {
			atomic.AddInt64(&poolQsortSz, 1)
			n.Tracef("poolQsortSz++ to %d\n", poolQsortSz)
		}
		return
	}

	mlo,mhi := doPivot(d, 0, l)
	c := 0
	xData := x.Data
	xName := x.Name
	x.Data = a.Data // recurse
	x.Name = x.Name+"("+a.Name+")"
	if mlo>0 {
		x.Val = x.AckWrap(a.Val.(Interface2).SubSlice(0, mlo))
		x.SendData(n)
		c++
	}
	if l-mhi>0 {
		x.Val = x.AckWrap(a.Val.(Interface2).SubSlice(mhi, l))
		x.SendData(n)
		c++
	}
	x.Data = xData
	x.Name = xName
	
	if steerAck {
		atomic.AddInt64(&poolQsortSz, 1)
		n.Tracef("poolQsortSz++ to %d\n", poolQsortSz)
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
		n[i] = MakeNode2("qsort", []*Edge{&aa}, []*Edge{&xx}, nil, qsortFire)
	}
	return n

}
