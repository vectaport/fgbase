package flowgraph

import (
)

func strval_func (n *Node) {
	a := n.Srcs[0]
	b := n.Srcs[1]
	x := n.Dsts[0]
	y := n.Dsts[1]
	x.Val = nil
	y.Val = nil
	if (ZeroTest(a.Val)) {
		x.Val = b.Val
	} else {
		y.Val = b.Val
	}
}

func strval_rdy (n *Node) bool {
	if n.Srcs[0].Rdy&&n.Srcs[1].Rdy {
		if ZeroTest(n.Srcs[0].Val) {
			return n.Dsts[0].Rdy
		} else {
			return n.Dsts[1].Rdy
		}
	} else {
		return false
	}
}

// Steer value goroutine
func FuncStrVal(a, b, x, y Edge) {

	node := MakeNode2("strval", []*Edge{&a, &b}, []*Edge{&x, &y}, strval_rdy, strval_func)
	node.Run()

}
