package flowgraph

import (
)

func const_func(n *Node) {
	x := n.Dsts[0]
	x.Val = x.Aux
}

// Constant value goroutine
func FuncConst(x Edge, c Datum) {

	node:=MakeNode("const", nil, []*Edge{&x}, nil, const_func)
	x.Aux = c
	node.Run()
}
