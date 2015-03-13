package flowgraph

import (
	"fmt"
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

func StrCndFunc(a, x, y Conn) {

	pipeid := MakePipe()

	var _a Datum = a.Init_val
	_a_rdy := a.Data_init
	_x_rdy := x.Ack_init
	_y_rdy := y.Ack_init

	for {
		fmt.Printf("	strcnd(%d):  _a_rdy %v  _x_rdy,_y_rdy %v,%v\n", pipeid, _a_rdy, _x_rdy, _y_rdy);

		if _a_rdy && _x_rdy && _y_rdy {
			fmt.Printf("	strcnd(%d):  writing x.Data or y.Data and a.Ack\n", pipeid)
			_a_rdy = false
			if (zerotest(_a)) {
				x.Data <- _a
				_x_rdy = false
			} else {
				y.Data <- _a
				_y_rdy = false
			}
			a.Ack <- true
			fmt.Printf("	strcnd(%d):  done writing x.Data or y.Data and a.Ack\n", pipeid)
		}

		fmt.Printf("	strcnd(%d):  select", pipeid)
		select {
		case _a = <-a.Data:
			{
				fmt.Printf("	strcnd(%d):  a read %v --  %v\n", pipeid, reflect.TypeOf(_a), _a)
				_a_rdy = true
			}
		case _x_rdy = <-x.Ack:
			fmt.Printf("	strcnd(%d):  x.Ack read\n", pipeid)
		case _y_rdy = <-y.Ack:
			fmt.Printf("	strcnd(%d):  y.Ack read\n", pipeid)
		}

	}

}
