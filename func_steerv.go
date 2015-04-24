package flowgraph

import (
)

func steervRdy (n *Node) bool {
	if n.Srcs[0].Rdy()&&n.Srcs[1].Rdy() {
		if ZeroTest(n.Srcs[0].Val) {
			return n.Dsts[0].Rdy()
		}
		return n.Dsts[1].Rdy()
	}
	return false
}

func steervFire (n *Node) {
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
func FuncSteerv(a, b, x, y Edge) {

	node := MakeNode("steerv", []*Edge{&a, &b}, []*Edge{&x, &y}, steervRdy, steervFire)
	node.Run()

}
