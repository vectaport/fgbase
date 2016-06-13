package flowgraph

import (
)

// FuncReduce reduces a stream of data into a single empty interface.
func FuncReduce(a,x Edge, reducer func(n *Node, datum,collection interface{}) interface{}) Node {

	var reduceFire = func (n *Node) {
		n.Aux = reducer(n, a.Val, n.Aux)
		x.Val = n.Aux
	}


	node := MakeNode("reduce", []*Edge{&a}, []*Edge{&x}, nil, reduceFire)
	node.Aux = make([]string, 0)
	return node
}
