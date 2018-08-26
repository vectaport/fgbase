package flowgraph

import (
	"sort"
)

// RecursiveSort extends sort.Interface for recursive sorting.
type RecursiveSort interface {
	sort.Interface

	// SubSlice returns a sub-slice.
	SubSlice(n, m int) interface{}

	// Slice returns current slice.
	Slice() []int
	// Original returns original slice
	Original() []int

	// Depth returns the depth of a recursive sort
	Depth() int64
	// ID returns a unique ID for the original slice.
	ID() int64
}

// FuncQsort recursively implements a quicksort with goroutines
// (x=qsort(a)).
func FuncQsort(a, x Edge, poolSz int) *Pool {

	var p *Pool

	qsortFire := func(n *Node) {

		a := n.Srcs[0]
		x := n.Dsts[0]

		// Ack early if Node available for upstream use.
		ackEarly := p.Alloc(n, 1)
		if ackEarly {
			a.Flow = true
			a.SendAck(n)
		}

		// If upstream is a Node from PoolQsort.
		recursed := n.Recursed()

		// Return the right number of Node's to the Pool.
		defer func() {
			m := 0
			if ackEarly {
				m++
			}
			if recursed {
				m++
			}
			if m != 0 {
				p.Free(n, m)
			}
		}()

		av := a.SrcGet()
		d, ok := av.(RecursiveSort)
		if !ok {
			n.LogError("not of type RecursiveSort (%T)\n", a.Val)
			return
		}

		n.Tracef("Original(%p) sorted %t, Sliced sorted %t, depth=%d, id=%d, len=%d, poolsz=%d\n",
			d.Original(), sort.IntsAreSorted(d.Original()), sort.IntsAreSorted(d.Slice()), d.Depth(), d.ID(), d.Len(), p.size)
		if d.Depth() == 0 {
			n.Tracef("BEGIN for id=%d, depth=0, len=%d\n", d.ID(), d.Len())
		}

		l := d.Len()

		if l <= ChannelSize || !p.Alloc(n, 2) {
			sort.Sort(d)
			x.DstPut(n.NodeWrap(d, x.Ack))
			x.SendData(n)
			return
		}

		mlo, mhi := doPivot(d, 0, l)

		// Make a substitute output Edge to point back to the Pool.
		xBack := x.PoolEdge(a)

		var lo, hi interface{}
		if mlo > 1 {
			n.Tracef("Original(%p) recurse left [0:%d], id=%d, depth will be %d\n", d.Original(), mlo, d.ID(), d.Depth()+1)
			lo = n.NodeWrap(d.SubSlice(0, mlo), x.Ack)
			xBack.DstPut(lo)
			xBack.SendData(n)
			x.RdyCnt++
		} else {
			p.Free(n, 1)
		}
		if l-mhi > 1 {
			n.Tracef("Original(%p) recurse right [%d:%d], id=%d, depth will be %d\n", d.Original(), mhi, l, d.ID(), d.Depth()+1)
			hi = n.NodeWrap(d.SubSlice(mhi, l), x.Ack)
			xBack.DstPut(hi)
			xBack.SendData(n)
			x.RdyCnt++
		} else {
			p.Free(n, 1)
		}

		x.Val = []interface{}{lo, hi} // for tracing as lo|hi.

	}

	// Make a pool of qsort nodes that can be dynamically used,
	recurse := true
	spread := false
	p = MakePool(poolSz, "qsort", []Edge{a}, []Edge{x}, nil, qsortFire, recurse, spread)
	return p

}
