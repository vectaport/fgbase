package flowgraph

import (
)

func rdyFire(n *Node) {
	a := n.Srcs[0]
	x := n.Dsts[0]
	x.Val = a.Val
}

// FuncRdy waits for two values before passing on the first (b; x = a).
func FuncRdy(a, b, x Edge) {

	node := MakeNode("rdy", []*Edge{&a, &b}, []*Edge{&x}, nil, rdyFire)
	node.Run()
}
