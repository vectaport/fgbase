package flowgraph


type Interface2 interface {
	// Interface is borrowed from the sort package.
	Interface
	// Sorted tests if array is sorted.
	Sorted() bool
}

func (n *Node) sortSz(l int, ch string) {
	if TraceLevel>=VVV {
		var s string
		for i :=0; i<(1024*1024); i += (1024*1024)/10 {
			if i<l {
				s += ch
			} else {
				break
			}
		}
		n.Tracef("%s\n", s)
	}
}

func qsortFire (n *Node) {
	a := n.Srcs[0]
	x := n.Dsts[0]
	a.SendAck(n) // write early to let flow go on
	a.NoOut = true
	if _,ok := a.Val.(Interface2); !ok {
		n.Errorf("not of type Interface2 (%T)\n", a.Val)
		return
	}

	l := Len(a.Val)
	n.sortSz(l, ">")
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
	n.sortSz(l, "<")
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
