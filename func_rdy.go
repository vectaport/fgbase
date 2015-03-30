package flowgraph

import (
)

// Ready (synchronization) goroutine
func FuncRdy(a, b, x Edge) {

	node := NewNode("rdy", []*Edge{&a, &b}, []*Edge{&x}, nil)

	for {
		if node.Rdy() {
			node.Tracef("writing x.Data and a.Ack and b.Ack\n")

			x.Val = a.Val
			node.TraceVals()

			x.Data <- x.Val
			a.Ack <- true
			b.Ack <- true
			node.Tracef("done writing x.Data and a.Ack and b.Ack\n")

			a.Rdy = false
			b.Rdy = false
			x.Rdy = false
		}

		node.Select()

	}

}
