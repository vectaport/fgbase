package flowgraph

func qsortFire (n *Node) {
	a := n.Srcs[0]
	x := n.Dsts[0]
	if _,ok := a.Val.(Interface); !ok {
		n.Errorf("not of type Interface (%T)\n", a.Val)
		return
	}
	l := Len(a.Val)
	d := a.Val.(Interface)
	if l <= 1024 {
		Sort(d)
	} else {
		n.Tracef("READY FOR SPECIAL SORT\n")
		mlo,mhi := doPivot(d, 0, l-1)
		maxDepth := 0
		for i := l; i > 0; i >>= 1 {
			maxDepth++
		}
		maxDepth *= 2
		maxDepth--
		quickSort(d, 0, mlo, maxDepth)
		quickSort(d, mhi, l-1, maxDepth)
	}
	x.Val=d
}

// FuncQsort recursively implements a quicksort with goroutines (x=qsort(a)).
func FuncQsort(a, x Edge) Node {

	node := MakeNode("qsort", []*Edge{&a}, []*Edge{&x}, nil, qsortFire)
	return node

}
