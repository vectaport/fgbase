package flowgraph

import (
)

func steervRdy (n *Node) bool {
	a := n.Srcs[0]
	b := n.Srcs[1]
	x := n.Dsts[0]
	y := n.Dsts[1]
	if a.Rdy()&&b.Rdy() {
		if ZeroTest(a.Val) {
			return x.Rdy()
		}
		return y.Rdy()
	}
	return false
}

func steervWork (n *Node) {
	a := n.Srcs[0]
	b := n.Srcs[1]
	x := n.Dsts[0]
	y := n.Dsts[1]
	if (ZeroTest(a.Val)) {
		x.Val = b.Val
		y.NoOut = true
	} else {
		y.Val = b.Val
		x.NoOut = true
	}
}

// FuncSteerv steers the second value by the first (if a==0 { x = b } else { y = b }).
func FuncSteerv(a, b, x, y Edge) Node {

	node := MakeNode("steerv", []*Edge{&a, &b}, []*Edge{&x, &y}, steervRdy, steervWork)
	return node

}
