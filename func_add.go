package flowgraph

import (
	"errors"
	"fmt"
	"reflect"
	"runtime"
)

func func_add(a, b Datum) Datum {
	
	switch a.(type) {
        case int8: { return a.(int8)+b.(int8) }
        case uint8: { return a.(uint8)+b.(uint8) }
        case int16: { return a.(int16)+b.(int16) }
        case uint16: { return a.(uint16)+b.(uint16) }
        case int32: { return a.(int32)+b.(int32) }
        case uint32: { return a.(uint32)+b.(uint32) }
        case int64: { return a.(int64)+b.(int64) }
        case uint64: { return a.(uint64)+b.(uint64) }
	case int: { return a.(int)+b.(int) }
	case uint: { return a.(uint)+b.(uint) }
	case float32: { return a.(float32)+b.(float32) }
	case float64: { return a.(float64)+b.(float64) }
	case complex64: { return a.(complex64)+b.(complex64) }
	case complex128: { return a.(complex128)+b.(complex128) }
	default: { return nil }
	}
}

// Addition goroutine
func FuncAdd(a, b, x Edge) {

	node := MakeNode("add", []*Edge{&a, &b}, []*Edge{&x}, nil)

	for {
		if node.Rdy() {

			atmp,btmp,same := Promote(a.Val, b.Val)

			if(!same) {
				_,nm,ln,_ := runtime.Caller(0)
				x.Val = errors.New(fmt.Sprintf("%s:%d (node.Id %d)  incompatible type for add operation (%v,%v)", nm, ln, node.Id, reflect.TypeOf(a), reflect.TypeOf(b)))
			} else {
				x.Val = func_add(atmp, btmp)
			}

			node.TraceVals()

			x.SendData()
			a.SendAck()
			b.SendAck()

		}

		node.Select()

	}

}
