package flowgraph

import (
)

// FuncFunc is the fully general func with any number of inputs and outputs,
// and use of a rather general function pointer that accepts and produces a slice
// of empty interfaces (Datum) (the *Node is for tracing).
func FuncFunc(src, dst []Edge, f func(*Node, []Datum) []Datum, anyRdy bool ) Node {

	var funcRdy = func(n *Node) bool {
		// if anyRdy { return true }
		return n.RdyAll()
	}

	var funcFire = func(n *Node) {
		var srcv []Datum
		for i := range n.Srcs {
			srcv = append(srcv, src[i].Val)
		}
		dstv := f(n, srcv)
		for i := range dstv {
			n.Dsts[i].Val = dstv[i]
		}
		
	}

	var srcp []*Edge
	for i := range src {
		srcp = append(srcp, &src[i])
	}
	var dstp []*Edge
	for i := range dst {
		dstp = append(dstp, &dst[i])
	}

	node := MakeNode("func", srcp, dstp, funcRdy, funcFire)
	return node
	
}
	
