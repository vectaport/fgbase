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
		n.Tracef("Sorted before?  %v\n", d.Sorted())
		Sort(d)
		n.Tracef("Sorted after?  %v\n", d.Sorted())
		x.Val=d
		return
	}
	n.Tracef("READY FOR SPECIAL SORT\n")
	mlo,mhi := doPivot(d, 0, l)
	n.Tracef("mlo,mhi %d,%d\n", mlo, mhi)
	maxDepth := 0
	for i := l; i > 0; i >>= 1 {
		maxDepth++
	}
	maxDepth *= 2
	maxDepth--
	n.Tracef("Sorted before?  %v\n", d.Sorted())
	quickSort(d, 0, mlo, maxDepth)
	quickSort(d, mhi, l, maxDepth)
	n.Tracef("Sorted after?  %v\n", d.Sorted())
	x.Val = d
}

// FuncQsort recursively implements a quicksort with goroutines (x=qsort(a)).
func FuncQsort(a, x Edge) Node {

	node := MakeNode("qsort", []*Edge{&a}, []*Edge{&x}, nil, qsortFire)
	return node

}
