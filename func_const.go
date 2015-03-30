package flowgraph

import (
)

// Constant value goroutine
func FuncConst(x Edge) {


	node:=NewNode("const", nil, []*Edge{&x}, nil)

	for {

		if node.Rdy() {
			node.TraceVals()
			node.Tracef("writing x.Data: %d\n", x.Val.(int))
			x.Data <- x.Val
			x.Rdy = false
		}

		node.Select()
	}
	
}
