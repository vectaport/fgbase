package flowgraph

import (
)

// Constant value goroutine
func FuncConst(x Edge) {


	node:=MakeNode("const", nil, []*Edge{&x}, nil)

	for {

		if node.Rdy() {
			node.TraceVals()
			if(x.Data!=nil) {x.Data <- x.Val; x.Rdy = false }
		}

		node.Select()
	}
	
}
