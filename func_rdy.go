package flowgraph

import (
)

// Ready (synchronization) goroutine
func FuncRdy(a, b, x Edge) {

	node := MakeNode("rdy", []*Edge{&a, &b}, []*Edge{&x}, nil)

	for {
		if node.Rdy() {
			node.Tracef("writing x.Data and a.Ack and b.Ack\n")

			x.Val = a.Val
			node.TraceVals()
			
			if (x.Data!= nil) {x.Data <- x.Val; x.Rdy = false}
			if (a.Ack!=nil) {a.Ack<- true; a.Rdy = false}
			if (b.Ack!=nil) {b.Ack<- true; b.Rdy = false}
		}

		node.Select()

	}

}
