package flowgraph

import (
)

func constFire(n *Node) {
	x := n.Dsts[0]
	x.Val = x.Aux
}

// FuncConst produces a constant value (x = c).
func FuncConst(x Edge, c Datum) {

	node:=MakeNode("const", nil, []*Edge{&x}, nil, constFire)
	x.Aux = c
	node.Run()
}
