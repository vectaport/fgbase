package flowgraph

import (
)

func constFire(n *Node) {
	x := n.Dsts[0]
	x.Val = n.Aux
}

// FuncConst produces a constant value (x = c).  Can also
// be done with an Edge made const.
func FuncConst(x Edge, c Datum) Node {

	node:=MakeNode("const", nil, []*Edge{&x}, nil, constFire)
	node.Aux = c
	return node
}
