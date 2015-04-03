package flowgraph

import (
)

func rdy_func(n *Node) {
	a := n.Srcs[0]
	x := n.Dsts[0]
	x.Val = a.Val
}

// Ready (synchronization) goroutine
func FuncRdy(a, b, x Edge) {

	node := MakeNode("rdy", []*Edge{&a, &b}, []*Edge{&x}, nil, rdy_func)
	node.Run()
}
