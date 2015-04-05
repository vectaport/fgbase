package flowgraph

import (
)

// FuncPass sends a value two ways (x = a)
func FuncPass(a, x Edge) {

	node := MakeNode("pass", []*Edge{&a}, []*Edge{&x}, nil, func(n *Node) {n.Dsts[0].Val = n.Srcs[0].Val} )
	node.Run()

}
