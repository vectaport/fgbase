package flowgraph

import (
	"reflect"
)

func RdyNode(a, b, x Edge) {

	node := MakeNode("rdy", []*Edge{&a, &b}, []*Edge{&x}, nil)

	for {
		if node.Rdy() {
			node.Printf("writing x.Data and a.Ack and b.Ack\n")

			x.Val = a.Val
			node.PrintVals()

			x.Data <- x.Val
			a.Ack <- true
			b.Ack <- true
			node.Printf("done writing x.Data and a.Ack and b.Ack\n")

			a.Rdy = false
			b.Rdy = false
			x.Rdy = false
		}

		node.Printf("select\n")
		select {
		case a.Val = <-a.Data:
			{
				node.Printf("a.Data read %v --  %v\n", reflect.TypeOf(a.Val), a.Val)
				a.Rdy = true
			}
		case b.Val = <-b.Data:
			{
				node.Printf("b.Data read %v --  %v\n", reflect.TypeOf(b.Val), b.Val)
				b.Rdy = true
			}
		case x.Rdy = <-x.Ack:
			node.Printf("x.Ack read\n")
		}

	}

}
