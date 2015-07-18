package flowgraph

import (
)

// FuncReduce reduces a stream of data into a single Datum.
func FuncReduce(a,x Edge, poolSz int, reducer func(i,d Datum)) Pool {

	i := 0
	var dict Datum
	var reduceFire = func (n *Node) {
		a := n.Srcs[0]
		reducer(a.Val, dict)
		i++
		if i%100==0 {
			x.Val = dict
		} else {
			x.NoOut = true
		}
	}

	// Make a pool of reduce nodes that share input and output channels
	recurse := false
	return MakePool(poolSz, "reduce", []Edge{a}, []Edge{x}, nil, reduceFire, recurse)

}
