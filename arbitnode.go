package flowgraph

import (
	"reflect"
)

func ArbitNode(a, b, x Edge) {

	node := MakeNode("arbit")

	var _a Datum = a.Init_val
	var _b Datum = a.Init_val
	_a_rdy := a.Data_init
	_b_rdy := b.Data_init
	_x_rdy := x.Ack_init

	_a_last := false

	for {
		node.Printf("_a_rdy,_b_rdy %v,%v  _x_rdy %v\n", _a_rdy, _b_rdy, _x_rdy);

		if (_a_rdy || _b_rdy) && _x_rdy {
			node.ExecCnt()
			node.Printf("writing x.Data  and either a.Ack or b.Ack\n")
			if(_a_rdy && !_b_rdy || _a_rdy && !_a_last) {
				_a_rdy = false
				_x_rdy = false
				_a_last = true
				x.Data <- _a
				a.Ack <- true
				node.Printf("done writing x.Data and a.Ack\n")
			} else if (_b_rdy) {
				_b_rdy = false
				_x_rdy = false
				_a_last = false
				x.Data <- _b
				b.Ack <- true
				node.Printf("done writing x.Data and b.Ack\n")
			}
		}

		node.Printf("select\n")
		select {
		case _a = <-a.Data:
			{
				node.Printf("a read %v --  %v\n", reflect.TypeOf(_a), _a)
				_a_rdy = true
			}
		case _b = <-b.Data:
			{
				node.Printf("b read %v --  %v\n", reflect.TypeOf(_b), _b)
				_b_rdy = true
			}
		case _x_rdy = <-x.Ack:
			node.Printf("x.Ack read\n")
		}

	}

}
