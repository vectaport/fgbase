package flowgraph

import (
	"sort"
)

type DoubleDatum struct {
	a,b Datum
}

var PoolQsort Pool

type RecursiveSort interface {
	sort.Interface

	// SubSlice returns a sub-slice.
	SubSlice(n, m int) Datum

	// Slice returns current slice.
	Slice() []int
	// SliceSorted tests if current slice is sorted.
	SliceSorted() bool

	// Original returns original slice
	Original() []int
	// OriginalSorted tests if original slice is sorted.
	OriginalSorted() bool

	// Depth returns the depth of a recursive sort
	Depth() int64
	// ID returns a unique ID for the object
	ID() int64
}

func qsortFire (n *Node) {

	a := n.Srcs[0]
	x := n.Dsts[0]

	p := &PoolQsort

	recursed := n.Recursed()

	// If you can reserve a Pool Node for the next upstream use then ack early.
	ackEarly := p.Alloc(n, 1)
	if ackEarly { 
		a.SendAck(n)
		a.NoOut = true
	}

	// Return the right number of nodes to the pool.
	defer func() {
		m := 0
		if ackEarly { m++  }
		if recursed { m++ }
		if m!=0 { p.Free(n, m) }
	}()

	d,ok := a.Val.(RecursiveSort)
	if !ok {
		n.LogError("not of type RecursiveSort (%T)\n", a.Val)
		return
	}

	if d.Depth()==0 { n.Tracef("BEGIN for id=%d, depth=0, len=%d\n", d.ID(), d.Len()) }
	n.Tracef("Original(%p) sorted %t, Sliced sorted %t, depth=%d, id=%d, len=%d, poolsz=%d\n", d.Original(), d.OriginalSorted(), d.SliceSorted(), d.Depth(), d.ID(), d.Len(), p.size )

	l := d.Len()

	if l <= 4096 || !p.Alloc(n, 2) {
		sort.Sort(d)
		x.Val=n.NodeWrap(d)
		x.SendData(n)
		x.NoOut = true
		return
	}

	mlo,mhi := doPivot(d, 0, l)
	var lo,hi Datum
	c := 0
	xData := x.Data
	xName := x.Name
	x.Data = a.Data // recurse
	x.Name = x.Name+"("+a.Name+")"
	if mlo>0 {
		n.Tracef("Original(%p) recurse left [0:%d], id=%d, depth will be %d\n", d.Original(), mlo, d.ID(), d.Depth()+1)
		d2 := d.SubSlice(0, mlo)
		lo = n.NodeWrap(d2)
		n.Tracef("Ack for left callback %p\n", n.Dsts[0].Ack)
		x.Val = lo
		x.SendData(n)
		c++
	} else {
		p.Free(n, 1)
	}
	if l-mhi>0 {
		n.Tracef("Original(%p) recurse right [%d:%d], id=%d, depth will be %d\n", d.Original(), mhi, l, d.ID(), d.Depth()+1)
		hi = n.NodeWrap(d.SubSlice(mhi, l))
		n.Tracef("Ack for right callback %p\n", n.Dsts[0].Ack)
		x.Val = hi
		x.SendData(n)
		c++
	} else {
		p.Free(n, 1)
	}
	x.Data = xData
	x.Name = xName

	x.Val = DoubleDatum{lo, hi}
	x.NoOut = true
	
	x.RdyCnt = c
	x.NoOut = true
	
}

// FuncQsort recursively implements a quicksort with goroutines 
// (x=qsort(a)).
func FuncQsort(a, x Edge, poolSz, poolReserve int ) *Pool {
	
	// Make a pool of qsort nodes that can be dynamically used, 
	PoolQsort = MakePool(poolSz, poolReserve, "qsort", []Edge{a}, []Edge{x}, nil, qsortFire)
	return &PoolQsort

}
