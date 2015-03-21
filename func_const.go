package flowgraph

import (
)

func FuncConst(x Edge) {


	node:=MakeNode("const", nil, []*Edge{&x}, nil)

	for {

		if node.Rdy() {
			node.PrintVals()
			node.Printf("writing x.Data: %d\n", x.Val.(int))
			x.Data <- x.Val
			x.Rdy = false
		}

		node.Printf("select\n")
		select {
		case x.Rdy = <-x.Ack:
			node.Printf("x.Ack read\n")
		}
	}
	
}
