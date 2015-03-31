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

	node := MakeNode("strcnd", []*Edge{&a}, []*Edge{&x, &y}, strcnd_rdy)

	for {
		node.Tracef("a.Rdy %v  x.Rdy,y.Rdy %v,%v\n", a.Rdy, x.Rdy, y.Rdy);

		if node.Rdy() {
			x.Val = nil
			y.Val = nil
			if (ZeroTest(a.Val)) {
				x.Val = a.Val
				node.TraceVals()
				if(x.Data != nil) {x.Data <- x.Val; x.Rdy = false}
				
			} else {
				y.Val = a.Val
				node.TraceVals()
				if(y.Data!=nil) {y.Data <- y.Val; y.Rdy = false}
			}
			if (a.Ack!=nil) { a.Ack <- true; a.Rdy = false}
		}

		node.Select()

	}

}
