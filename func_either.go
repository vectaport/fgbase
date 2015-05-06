package flowgraph

import (
)

func eitherWork (n *Node) {
	a := n.Srcs[0]
	b := n.Srcs[1]
	x := n.Dsts[0]
	if a.Rdy() {
		x.Val = a.Val
		b.NoOut = true
	} else {
		x.Val = b.Val
		a.NoOut = true
	}
}

func eitherRdy (n *Node) bool {
	return (n.Srcs[0].Rdy() || n.Srcs[1].Rdy()) && n.Dsts[0].Rdy()
}

// FuncEither passes on one of two values (if =a { x = a } else { x = b }).
func FuncEither(a, b, x Edge) Node {

	node := MakeNode("either", []*Edge{&a, &b}, []*Edge{&x}, eitherRdy, eitherWork)
	return node

}
