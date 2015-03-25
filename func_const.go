package flowgraph

import (
)

// Constant value goroutine
func FuncConst(x Edge) {


	node:=NewNode("const", nil, []*Edge{&x}, nil)

	for {

		if node.Rdy() {
			node.TraceVal()
			node.Tracef("writing x.Data: %d\n", x.Val.(int))
			x.Data <- x.Val
			x.Rdy = false
		}

		node.Tracef("select\n")
		select {
		case x.Rdy = <-x.Ack:
			node.Tracef("x.Ack read\n")
		}
	}
	
}
