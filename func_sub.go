package flowgraph

import (
	"errors"
	"fmt"
	"reflect"
	"runtime"
)

func func_sub(a, b Datum) Datum {
	
	switch a.(type) {
        case int8: { return a.(int8)-b.(int8) }
        case uint8: { return a.(uint8)-b.(uint8) }
        case int16: { return a.(int16)-b.(int16) }
        case uint16: { return a.(uint16)-b.(uint16) }
        case int32: { return a.(int32)-b.(int32) }
        case uint32: { return a.(uint32)-b.(uint32) }
        case int64: { return a.(int64)-b.(int64) }
        case uint64: { return a.(uint64)-b.(uint64) }
	case int: { return a.(int)-b.(int) }
	case uint: { return a.(uint)-b.(uint) }
	case float32: { return a.(float32)-b.(float32) }
	case float64: { return a.(float64)-b.(float64) }
	case complex64: { return a.(complex64)-b.(complex64) }
	case complex128: { return a.(complex128)-b.(complex128) }
	default: { return nil }
	}
}

// subtraction goroutine
func FuncSub(a, b, x Edge) {

	node := NewNode("sub", []*Edge{&a, &b}, []*Edge{&x}, nil)

	for {
		node.Tracef("a.Rdy,b.Rdy %v,%v  x.Rdy %v\n", a.Rdy, b.Rdy, x.Rdy);

		if node.Rdy() {
			node.Tracef("writing x.Data and a.Ack and b.Ack\n")

			if(reflect.TypeOf(a.Val)!=reflect.TypeOf(b.Val)) {
				_,nm,ln,_ := runtime.Caller(0)
				x.Val = errors.New(fmt.Sprintf("%s:%d (node.Id %d)  type mismatch (%v,%v)", nm, ln, node.Id, reflect.TypeOf(a.Val), reflect.TypeOf(b.Val)))
			} else {
				x.Val = func_sub(a.Val, b.Val)
			}
			node.PrintVals()
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
