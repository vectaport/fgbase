package flowgraph

import (
)

// FuncReduce reduces a stream of data into a single Datum.
func FuncReduce(a,x Edge, reducer func(n *Node, s,d Datum) Datum) Node {

	i := 0
	var reduceFire = func (n *Node) {
		a := n.Srcs[0]
		a.Aux = reducer(n, a.Val, a.Aux)
		i++
		if i%1000==0 || true {
			x.Val = a.Aux
		} else {
			x.NoOut = true
		}
	}

	a.Aux = make([]string, 0)

	node := MakeNode("reduce", []*Edge{&a}, []*Edge{&x}, nil, reduceFire)
	return node
}
