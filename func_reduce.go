package flowgraph

import (
)

// FuncReduce reduces a stream of data into a single Datum.
func FuncReduce(a,x Edge, reducer func(n *Node, datum,collection Datum) Datum) Node {

	var reduceFire = func (n *Node) {
		a.Aux = reducer(n, a.Val, a.Aux)
		x.Val = a.Aux
	}

	a.Aux = make([]string, 0)

	node := MakeNode("reduce", []*Edge{&a}, []*Edge{&x}, nil, reduceFire)
	return node
}
