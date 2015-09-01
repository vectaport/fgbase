package flowgraph

import (
)

// FuncFunc is the fully general func with any number of input and output Edge's,
// and fully general ready and fire funcs.
func FuncFunc(name string, src, dst []*Edge, rdyFunc NodeRdy, fireFunc NodeFire) Node {

	node := MakeNode(name, src, dst, rdyFunc, fireFunc)
	return node
	
}
	
