package flowgraph

import (
)

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

	node := MakeNode("strval", []*Edge{&a, &b}, []*Edge{&x, &y}, strval_rdy)

	for {

		if node.Rdy() {
			x.Val = nil
			y.Val = nil
			if (ZeroTest(a.Val)) {
				x.Val = b.Val
				node.TraceVals()
				if (x.Data != nil) {x.Data <- x.Val; x.Rdy = false}
				
			} else {
				y.Val = b.Val
				node.TraceVals()
				if (y.Data != nil) {y.Data <- y.Val; y.Rdy = false}
			}
			if (a.Ack!=nil) {a.Ack <- true; a.Rdy = false}
			if (b.Ack!=nil) {b.Ack <- true; b.Rdy = false}
		}

		node.Select()

	}

}
