package flowgraph

import (
	"sort"
	"sync"
	"sync/atomic"
)

// These are sort.min, sort.medianOfThree, sort.swapRange, and sort.doPivot, borrowed 
// under the GO-LICENSE from go 1.4.2.  Required to reuse sort.doPivot.
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// medianOfThree moves the median of the three values data[a], data[b], data[c] into data[a].
func medianOfThree(data sort.Interface, a, b, c int) {
	m0 := b
	m1 := a
	m2 := c
	// bubble sort on 3 elements
	if data.Less(m1, m0) {
		data.Swap(m1, m0)
	}
	if data.Less(m2, m1) {
		data.Swap(m2, m1)
	}
	if data.Less(m1, m0) {
		data.Swap(m1, m0)
	}
	// now data[m0] <= data[m1] <= data[m2]
}

func swapRange(data sort.Interface, a, b, n int) {
	for i := 0; i < n; i++ {
		data.Swap(a+i, b+i)
	}
}

func doPivot(data sort.Interface, lo, hi int) (midlo, midhi int) {
	m := lo + (hi-lo)/2 // Written like this to avoid integer overflow.
	if hi-lo > 40 {
		// Tukey's ``Ninther,'' median of three medians of three.
		s := (hi - lo) / 8
		medianOfThree(data, lo, lo+s, lo+2*s)
		medianOfThree(data, m, m-s, m+s)
		medianOfThree(data, hi-1, hi-1-s, hi-1-2*s)
	}
	medianOfThree(data, lo, m, hi-1)

	// Invariants are:
	//	data[lo] = pivot (set up by ChoosePivot)
	//	data[lo <= i < a] = pivot
	//	data[a <= i < b] < pivot
	//	data[b <= i < c] is unexamined
	//	data[c <= i < d] > pivot
	//	data[d <= i < hi] = pivot
	//
	// Once b meets c, can swap the "= pivot" sections
	// into the middle of the slice.
	pivot := lo
	a, b, c, d := lo+1, lo+1, hi, hi
	for {
		for b < c {
			if data.Less(b, pivot) { // data[b] < pivot
				b++
			} else if !data.Less(pivot, b) { // data[b] = pivot
				data.Swap(a, b)
				a++
				b++
			} else {
				break
			}
		}
		for b < c {
			if data.Less(pivot, c-1) { // data[c-1] > pivot
				c--
			} else if !data.Less(c-1, pivot) { // data[c-1] = pivot
				data.Swap(c-1, d-1)
				c--
				d--
			} else {
				break
			}
		}
		if b >= c {
			break
		}
		// data[b] > pivot; data[c-1] < pivot
		data.Swap(b, c-1)
		b++
		c--
	}

	n := min(b-a, a-lo)
	swapRange(data, lo, b-n, n)

	n = min(hi-d, d-c)
	swapRange(data, c, hi-n, n)

	return lo + b - a, hi - (d - c)
}

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
