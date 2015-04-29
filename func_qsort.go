package flowgraph


type Interface2 interface {
	// Interface is borrowed from the sort package.
	Interface
	// Sorted tests if array is sorted.
	Sorted() bool
}


func qsortFire (n *Node) {
	a := n.Srcs[0]
	x := n.Dsts[0]
	if _,ok := a.Val.(Interface2); !ok {
		n.Errorf("not of type Interface2 (%T)\n", a.Val)
		return
	}

	l := Len(a.Val)
	d := a.Val.(Interface2)
	if l <= 1024 {
		Sort(d)
		x.Val=x.AckWrap(d)
		return
	}

	mlo,mhi := doPivot(d, 0, l)
	maxDepth := 0
	for i := l; i > 0; i >>= 1 {
		maxDepth++
	}
	maxDepth *= 2
	maxDepth--
	quickSort(d, 0, mlo, maxDepth)
	quickSort(d, mhi, l, maxDepth)
	x.Val = x.AckWrap(d)
}

// FuncQsort recursively implements a quicksort with goroutines (x=qsort(a)).
func FuncQsort(a, x Edge, poolSz int) []Node {
	
	n := MakeNodes(poolSz)
	for i:=0; i<poolSz; i++ {
		aa, xx := a,x  // make a copy of the Edge's for each one
		n[i] = MakeNode2("qsort", []*Edge{&aa}, []*Edge{&xx}, nil, qsortFire)
	}
	return n

}
