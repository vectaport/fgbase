package flowgraph

import (
	"fmt"
	"reflect"
)

func ArbitNode(a, b, x Edge) {

	nodeid := MakeNode()

	var _a Datum = a.Init_val
	var _b Datum = a.Init_val
	_a_rdy := a.Data_init
	_b_rdy := b.Data_init
	_x_rdy := x.Ack_init

	_a_last := false

	for {
		fmt.Printf("	arbit(%d):  _a_rdy,_b_rdy %v,%v  _x_rdy %v\n", nodeid, _a_rdy, _b_rdy, _x_rdy);

		if (_a_rdy || _b_rdy) && _x_rdy {
			fmt.Printf("	arbit(%d):  writing x.Data  and either a.Ack or b.Ack\n", nodeid)
			if(_a_rdy && !_b_rdy || _a_rdy && !_a_last) {
				_a_rdy = false
				_x_rdy = false
				_a_last = true
				x.Data <- _a
				a.Ack <- true
				fmt.Printf("	arbit(%d):  done writing x.Data and a.Ack\n", nodeid)
			} else if (_b_rdy) {
				_b_rdy = false
				_x_rdy = false
				_a_last = false
				x.Data <- _b
				b.Ack <- true
				fmt.Printf("	arbit(%d):  done writing x.Data and b.Ack\n", nodeid)
			}
		}

		fmt.Printf("	arbit(%d):  select\n", nodeid)
		select {
		case _a = <-a.Data:
			{
				fmt.Printf("	arbit(%d):  a read %v --  %v\n", nodeid, reflect.TypeOf(_a), _a)
				_a_rdy = true
			}
		case _b = <-b.Data:
			{
				fmt.Printf("	arbit(%d):  b read %v --  %v\n", nodeid, reflect.TypeOf(_b), _b)
				_b_rdy = true
			}
		case _x_rdy = <-x.Ack:
			fmt.Printf("	arbit(%d):  x.Ack read\n", nodeid)
		}

	}

}
