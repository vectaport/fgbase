package flowgraph

import (
	"reflect"
)

func strval_rdy (n *Node) bool {
	if n.Srcs[0].Rdy&&n.Srcs[1].Rdy {
		if ZeroTest(n.Srcs[0].Val) {
			return n.Dsts[0].Rdy
		} else {
			return n.Dsts[1].Rdy
		}
	} else {
		return false
	}
}

// steer value goroutine
func FuncStrVal(a, b, x, y Edge) {

	node := NewNode("strval", []*Edge{&a, &b}, []*Edge{&x, &y}, strval_rdy)

	for {
		node.Tracef("a.Rdy,b.Rdy %v,%v  x.Rdy,y.Rdy %v,%v\n", a.Rdy, b.Rdy, x.Rdy, y.Rdy);

		if node.Rdy() {
			node.Tracef("writing x.Data or y.Data and a.Ack\n")
			x.Val = nil
			y.Val = nil
			if (ZeroTest(a.Val)) {
				node.Tracef("x write\n")
				x.Val = b.Val
				node.PrintVals()
				x.Data <- x.Val
				x.Rdy = false
				
			} else {
				node.Tracef("y write\n")
				y.Val = b.Val
				node.PrintVals()
				y.Data <- y.Val
				y.Rdy = false
			}
			a.Ack <- true
			b.Ack <- true
			a.Rdy = false
			b.Rdy = false
			node.Tracef("done writing x.Data or y.Data and a.Ack and b.Ack\n")
		}

		node.Tracef("select\n")
		select {
		case a.Val = <-a.Data:
			{
				node.Tracef("a read %v --  %v\n", reflect.TypeOf(a.Val), a.Val)
				a.Rdy = true
			}
		case b.Val = <-b.Data:
			{
				node.Tracef("a read %v --  %v\n", reflect.TypeOf(b.Val), b.Val)
				b.Rdy = true
			}
		case x.Rdy = <-x.Ack:
			node.Tracef("x.Ack read\n")
		case y.Rdy = <-y.Ack:
			node.Tracef("y.Ack read\n")
		}

	}

}
