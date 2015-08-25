package flowgraph

import (
)

// FuncFunc is the fully general func with any number of inputs and outputs,
// and fully general ready and fire funcs.
func FuncFunc(src, dst []Edge, rdyFunc NodeRdy, fireFunc NodeFire) Node {

	var srcp []*Edge
	for i := range src {
		srcp = append(srcp, &src[i])
	}
	var dstp []*Edge
	for i := range dst {
		dstp = append(dstp, &dst[i])
	}

	node := MakeNode("func", srcp, dstp, rdyFunc, fireFunc)
	return node
	
}
	
