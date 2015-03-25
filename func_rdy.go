package flowgraph

import (
	"reflect"
)

// Ready (synchronization) goroutine
func FuncRdy(a, b, x Edge) {

	node := NewNode("rdy", []*Edge{&a, &b}, []*Edge{&x}, nil)

	for {
		if node.Rdy() {
			node.Tracef("writing x.Data and a.Ack and b.Ack\n")

			x.Val = a.Val
			node.TraceVal()

			x.Data <- x.Val
			a.Ack <- true
			b.Ack <- true
			node.Tracef("done writing x.Data and a.Ack and b.Ack\n")

			a.Rdy = false
			b.Rdy = false
			x.Rdy = false
		}

		node.Tracef("select\n")
		select {
		case a.Val = <-a.Data:
			{
				node.Tracef("a.Data read %v --  %v\n", reflect.TypeOf(a.Val), a.Val)
				a.Rdy = true
			}
		case b.Val = <-b.Data:
			{
				node.Tracef("b.Data read %v --  %v\n", reflect.TypeOf(b.Val), b.Val)
				b.Rdy = true
			}
		case x.Rdy = <-x.Ack:
			node.Tracef("x.Ack read\n")
		}

	}

}
