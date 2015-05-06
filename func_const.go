package flowgraph

import (
)

func constWork(n *Node) {
	x := n.Dsts[0]
	x.Val = x.Aux
}

// FuncConst produces a constant value (x = c).
func FuncConst(x Edge, c Datum) Node {

	node:=MakeNode("const", nil, []*Edge{&x}, nil, constWork)
	x.Aux = c
	return node
}
