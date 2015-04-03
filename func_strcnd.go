package flowgraph

import (
)

func strcndFire (n *Node) {
	a := n.Srcs[0]
	x := n.Dsts[0]
	y := n.Dsts[1]
	x.Val = nil
	y.Val = nil
	if (ZeroTest(a.Val)) {
		x.Val = a.Val
	} else {
		y.Val = a.Val
	}
}

func strcndRdy (n *Node) bool {
	a := n.Srcs[0]
	x := n.Dsts[0]
	y := n.Dsts[1]
	if a.Rdy {
		if ZeroTest(a.Val) {
			return x.Rdy
		}
		return y.Rdy
	}
	return false
}

// FuncStrCnd steers a condition one of two ways (if !a { x = a } else { y = a }).
func FuncStrCnd(a, x, y Edge) {

	node := MakeNode("strcnd", []*Edge{&a}, []*Edge{&x, &y}, strcndRdy, strcndFire)
	node.Run()

}
