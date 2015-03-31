package flowgraph

import (
)

func arbit_rdy (n *Node) bool {
	return (n.Srcs[0].Rdy || n.Srcs[1].Rdy) && n.Dsts[0].Rdy
}

// Arbiter goroutine
func FuncArbit(a, b, x Edge) {

	node := MakeNode("arbit", []*Edge{&a, &b}, []*Edge{&x}, arbit_rdy)

	a_last := false

	for {
		node.Tracef("a.Rdy,b.Rdy %v,%v  x.Rdy %v\n", a.Rdy, b.Rdy, x.Rdy);

		if node.Rdy() {
			if(a.Rdy && !b.Rdy || a.Rdy && !a_last) {
				a_last = true
				x.Val = a.Val
				node.TraceVals()
				if (a.Ack!=nil) { a.Ack <- true; a.Rdy = false }
			} else if (b.Rdy) {
				a_last = false
				x.Val = b.Val
				node.TraceVals()
				if (b.Ack!=nil) { b.Ack <- true; b.Rdy = false }
			}
			if(x.Data!=nil) { x.Data <- x.Val; x.Rdy = false }
		}

		node.Select()

	}

}
