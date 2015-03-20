package flowgraph

import (
	"reflect"
)

func ArbitNode(a, b, x Edge) {

	node := MakeNode("arbit", []*Edge{&a, &b}, []*Edge{&x}, nil)

	a_last := false

	for {
		node.Printf("a.Rdy,b.Rdy %v,%v  x.Rdy %v\n", a.Rdy, b.Rdy, x.Rdy);

		if (a.Rdy || b.Rdy) && x.Rdy {
			node.ExecCnt()
			node.Printf("writing x.Data  and either a.Ack or b.Ack\n")
			if(a.Rdy && !b.Rdy || a.Rdy && !a_last) {
				a_last = true
				x.Val = a.Val
				node.PrintVals()
				a.Rdy = false
				a.Ack <- true
				node.Printf("done writing x.Data and a.Ack\n")
			} else if (b.Rdy) {
				a_last = false
				x.Val = b.Val
				node.PrintVals()
				b.Rdy = false
				b.Ack <- true
				node.Printf("done writing x.Data and b.Ack\n")
			}
			x.Data <- x.Val
			x.Rdy = false
		}

		node.Printf("select\n")
		select {
		case a.Val = <-a.Data:
			{
				node.Printf("a read %v --  %v\n", reflect.TypeOf(a.Val), a.Val)
				a.Rdy = true
			}
		case b.Val = <-b.Data:
			{
				node.Printf("b read %v --  %v\n", reflect.TypeOf(b.Val), b.Val)
				b.Rdy = true
			}
		case x.Rdy = <-x.Ack:
			node.Printf("x.Ack read\n")
		}

	}

}
