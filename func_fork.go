package flowgraph

import (
)

// FuncFork sends a value two ways (x = a; y = a)
func FuncFork(a, x, y Edge) {

	node := MakeNode("fork", []*Edge{&a}, []*Edge{&x, &y}, nil, func (n *Node) { x.Val = a.Val; y.Val = a.Val })
	node.Run()

}
