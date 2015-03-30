package flowgraph

import (
)

func strcnd_rdy (n *Node) bool {
	if n.Srcs[0].Rdy {
		if ZeroTest(n.Srcs[0].Val) {
			return n.Dsts[0].Rdy
		} else {
			return n.Dsts[1].Rdy
		}
	} else {
		return false
	}
}

// Steer condition goroutine
func FuncStrCnd(a, x, y Edge) {

	node := NewNode("strcnd", []*Edge{&a}, []*Edge{&x, &y}, strcnd_rdy)

	for {
		node.Tracef("a.Rdy %v  x.Rdy,y.Rdy %v,%v\n", a.Rdy, x.Rdy, y.Rdy);

		if node.Rdy() {
			node.Tracef("writing x.Data or y.Data and a.Ack\n")
			x.Val = nil
			y.Val = nil
			if (ZeroTest(a.Val)) {
				node.Tracef("x write\n")
				x.Val = a.Val
				node.TraceVals()
				x.Data <- x.Val
				x.Rdy = false
				
			} else {
				node.Tracef("y write\n")
				y.Val = a.Val
				node.TraceVals()
				y.Data <- y.Val
				y.Rdy = false
			}
			a.Rdy = false
			a.Ack <- true
			node.Tracef("done writing x.Data or y.Data and a.Ack\n")
		}

		node.Select()

	}

}
