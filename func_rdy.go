package flowgraph

import (
)

func rdyWork(n *Node) {
	a := n.Srcs[0]
	x := n.Dsts[0]
	x.Val = a.Val
}

// FuncRdy waits for two values before passing on the first (b; x = a).
func FuncRdy(a, b, x Edge) Node {

	node := MakeNode("rdy", []*Edge{&a, &b}, []*Edge{&x}, nil, rdyWork)
	return node
}
