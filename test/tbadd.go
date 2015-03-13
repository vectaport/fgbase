package main

import (
	"github.com/vectaport/flowgraph"
	"fmt"
	"math"
	"reflect"
	"time"
)

func tbi(a, b flowgraph.Conn) {

	flowgraph.MakePipe()

	var _a flowgraph.Datum = 0
	var _b flowgraph.Datum = 0
	fmt.Printf("%v,%v --> ", reflect.TypeOf(_a), reflect.TypeOf(_b))
	_a_rdy := a.Ack_init
	_b_rdy := b.Ack_init

	var _i int = 0
	for {
		if (_i>10) { break }

		if _a_rdy && _b_rdy{
			fmt.Printf("tbi:  writing a and b: %d,%d\n", _a.(int), _b.(int))
			_a_rdy = false
			_b_rdy = false
			a.Data <- _a
			b.Data <- _b
			_a = _a.(int) + 1
			_b = _b.(int) + 1
			_i = _i + 1
		}

		fmt.Println("tbi:  select")
		select {
		case _a_rdy = <-a.Ack:
			fmt.Println("tbi:  a_req read")

		case _b_rdy = <-b.Ack:
			fmt.Println("tbi:  b_req read")
		}

	}

	_a = float32(0)
	_b = float32(0)
	_i = 0

	for {
		if (_i>9) { break }
		if _a_rdy && _b_rdy{
			fmt.Printf("tbi:  writing a and b: %f,%f\n", _a.(float32), _b.(float32))
			_a_rdy = false
			_b_rdy = false
			a.Data <- _a
			b.Data <- _b
			_a = _a.(float32) + 1.
			_b = _b.(float32) + 1.
			_i = _i + 1
		}

		fmt.Println("tbi:  select")
		select {
		case _a_rdy = <-a.Ack:
			fmt.Println("tbi:  a_req read")

		case _b_rdy = <-b.Ack:
			fmt.Println("tbi:  b_req read")
		}

	}

	_a = uint64(math.MaxUint64)
	_b = -1
	_i = 0

	for {
		if (_i > 0) { break }

		if _a_rdy && _b_rdy{
			fmt.Printf("tbi:  writing a and b: %v,%v\n", _a, _b)
			_a_rdy = false
			_b_rdy = false
			a.Data <- _a
			b.Data <- _b
			// _a = _a.(float32) + 1.
			// _b = _b.(float32) + 1.
			_i = _i + 1
		}

		fmt.Println("tbi:  select")
		select {
		case _a_rdy = <-a.Ack:
			fmt.Println("tbi:  a_req read")

		case _b_rdy = <-b.Ack:
			fmt.Println("tbi:  b_req read")
		}

	}

	_a = int8(0)
	_b = uint64(0)
	_i = 0

	for  {
		if (_i > 0) { break }

		if _a_rdy && _b_rdy{
			fmt.Printf("tbi:  writing a and b: %v,%v\n", _a, _b)
			_a_rdy = false
			_b_rdy = false
			a.Data <- _a
			b.Data <- _b
			// _a = _a.(float32) + 1.
			// _b = _b.(float32) + 1.
			_i = _i + 1
		}

		fmt.Println("tbi:  select")
		select {
		case _a_rdy = <-a.Ack:
			fmt.Println("tbi:  a_req read")

		case _b_rdy = <-b.Ack:
			fmt.Println("tbi:  b_req read")
		}

	}

}

func tbo(x flowgraph.Conn) {

	flowgraph.MakePipe()

	var _x flowgraph.Datum
	_x_rdy := x.Data_init

	for {
		// fmt.Println("		tbo:  _x_rdy", _x_rdy)
		if _x_rdy {
			fmt.Println("		tbo:  writing x_req")
			x.Ack <- true
			_x_rdy = false
		}

		fmt.Println("		tbo:  select")
		select {
		case _x = <-x.Data:
			{
				fmt.Printf("		tbo:  x read %v --  %v\n", reflect.TypeOf(_x), _x)
				_x_rdy = true
			}
		}

	}

}

func main() {

	a := flowgraph.MakeConn(false,true,nil)
	b := flowgraph.MakeConn(false,true,nil)
	x := flowgraph.MakeConn(false,true,nil)

	go tbi(a, b)
	go flowgraph.AddFunc(a, b, x)
	go tbo(x)

	time.Sleep(1000000000)

}

