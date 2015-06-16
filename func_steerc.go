package flowgraph

import (
)

func steercFire (n *Node) {
	a := n.Srcs[0]
	x := n.Dsts[0]
	y := n.Dsts[1]
	x.Val = nil
	y.Val = nil
	if (ZeroTest(a.Val)) {
		x.Val = a.Val
		y.NoOut = true
	} else {
		y.Val = a.Val
		x.NoOut = true
	}
}

func steercRdy (n *Node) bool {
	a := n.Srcs[0]
	x := n.Dsts[0]
	y := n.Dsts[1]
	if a.SrcRdy(n) {
		if ZeroTest(a.Val) {
			return x.DstRdy(n)
		}
		return y.DstRdy(n)
	}
	return false
}

// FuncSteerc steers a condition one of two ways (if a==0 { x = a } else { y = a }).
func FuncSteerc(a, x, y Edge) Node {

	node := MakeNode("steerc", []*Edge{&a}, []*Edge{&x, &y}, steercRdy, steercFire)
	return node

}
