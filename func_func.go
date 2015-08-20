package flowgraph

import (
)

// FuncFunc is the fully general func with any number of inputs and outputs,
// and the fully arbitrary function pointer that accepts and produces a slice
// of empty interfaces (Datum).
func FuncFunc(src, dst []Edge, f func(*Node, []Datum) []Datum ) Node {
	var funcFire = func(n *Node) {
		srcv := make([]Datum, 0)
		for i := range n.Srcs {
			srcv = append(srcv, src[i].Val)
		}
		dstv := f(n, srcv)
		for i := range n.Dsts {
			n.Dsts[i].Val = dstv[i]
		}
		
	}


	srcp := make([]*Edge, 0)
	for i := range src {
		srcp = append(srcp, &src[i])
	}
	dstp := make([]*Edge, 0)
	for i := range dst {
		dstp = append(dstp, &dst[i])
	}

	node := MakeNode("func", srcp, dstp, nil, funcFire)
	return node
	
}
	
