package flowgraph

import (
	"reflect"
)

func zerotest(a Datum) bool {
	
	switch a.(type) {
        case int8: { return a.(int8)==0 }
        case uint8: { return a.(uint8)==0 }
        case int16: { return a.(int16)==0 }
        case uint16: { return a.(uint16)==0 }
        case int32: { return a.(int32)==0 }
        case uint32: { return a.(uint32)==0 }
	case int64: { return a.(int64)==0 }
        case uint64: { return a.(uint64)==0 }
	case int: { return a.(int)==0 }
	case uint: { return a.(uint)==0 }
	case float32: { return a.(float32)==0.0 }
	case float64: { return a.(float64)==0.0 }
	case complex64: { return a.(complex64)==0.0+0.0i }
	case complex128: { return a.(complex128)==0.0+0.0i }
	default: { return true }
	}
}

func StrCndNode(a, x, y Edge) {

	node := MakeNode("strcnd", []*Edge{&a}, []*Edge{&x, &y}, nil)

	for {
		node.Printf("a.Rdy %v  x.Rdy,y.Rdy %v,%v\n", a.Rdy, x.Rdy, y.Rdy);

		if node.Rdy() {
			node.Printf("writing x.Data or y.Data and a.Ack\n")
			x.Val = nil
			y.Val = nil
			if (zerotest(a.Val)) {
				node.Printf("x write\n")
				x.Val = a.Val
				node.PrintVals()
				x.Data <- x.Val
				x.Rdy = false
				
			} else {
				node.Printf("y write\n")
				y.Val = a.Val
				node.PrintVals()
				y.Data <- y.Val
				y.Rdy = false
			}
			a.Rdy = false
			a.Ack <- true
			node.Printf("done writing x.Data or y.Data and a.Ack\n")
		}

		node.Printf("select\n")
		select {
		case a.Val = <-a.Data:
			{
				node.Printf("a read %v --  %v\n", reflect.TypeOf(a.Val), a.Val)
				a.Rdy = true
			}
		case x.Rdy = <-x.Ack:
			node.Printf("x.Ack read\n")
		case y.Rdy = <-y.Ack:
			node.Printf("y.Ack read\n")
		}

	}

}
