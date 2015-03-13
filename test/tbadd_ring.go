package main

import (
	"github.com/vectaport/flowgraph"
	"fmt"
	"reflect"
	"time"
)

func tbi(g, a, b flowgraph.Conn) {

	pipeid := flowgraph.MakePipe()
	
	var _g flowgraph.Datum
	var _a flowgraph.Datum = a.Init_val
	var _b flowgraph.Datum = b.Init_val
	
	_g_rdy := g.Data_init
	_a_rdy := a.Ack_init
	_b_rdy := b.Ack_init
	
	for {
		fmt.Printf("tbi(%d):  _g_rdy %v, _a_rdy,_b_rdy %v,%v\n", pipeid, _g_rdy, _a_rdy, _b_rdy);
		
		if _a_rdy && _b_rdy && _g_rdy {
			//fmt.Printf("tbi(%d)  writing a and b and g_req: %d,%d\n", pipeid, _a.(int), _b.(int))
			_a_rdy = false
			_b_rdy = false
			_g_rdy = false
			g.Ack <- true
			fmt.Printf("tbi(%d)  g.Ack written\n", pipeid);
			a.Data <- _a
			fmt.Printf("tbi(%d)  a.Data written\n", pipeid);
			b.Data <- _b
			fmt.Printf("tbi(%d)  b.Data written\n", pipeid);
			_a = _a.(int) + 1
			_b = _b.(int) + 1
		}
		
		fmt.Printf("tbi(%d)  select\n", pipeid)
		select {
		case _a_rdy = <-a.Ack: {
			fmt.Printf("tbi(%d)  a.Ack read\n", pipeid)
		}
			
		case _b_rdy = <-b.Ack: {
			fmt.Printf("tbi(%d)  b.Ack read\n", pipeid)
		}
			
		case _g = <-g.Data: {
			fmt.Printf("tbi(%d)  g.Data read\n", pipeid)
			flowgraph.Sink(_g)
			_g_rdy = true
		}
		}
		
	}
}

func tbo(x, g flowgraph.Conn) {

	pipeid := flowgraph.MakePipe()

	var _x flowgraph.Datum
	_x_rdy := x.Data_init
	_g_rdy := g.Ack_init

	for {
		fmt.Printf("		tbo(%d):  _x_rdy %v, _g_rdy %v\n", pipeid, _x_rdy, _g_rdy);
		if _x_rdy && _g_rdy {
			fmt.Printf("		tbo(%d):  writing g.Data and x.Ack\n", pipeid)
			g.Data <- true
			fmt.Printf("		tbo(%d):  done writing g.Data\n", pipeid)
			x.Ack <- true
			fmt.Printf("		tbo(%d):  done writing x.Ack\n", pipeid)
			_x_rdy = false
			_g_rdy = false
		}

		fmt.Printf("		tbo(%d):  select\n", pipeid)
		select {
		case _x = <-x.Data:
			{
				fmt.Printf("		tbo(%d):  x read %v --  %v\n", pipeid, reflect.TypeOf(_x), _x)
				_x_rdy = true
			}
		case _g_rdy = <-g.Ack:
			fmt.Println("		tbo(%d):  g.Ack read", pipeid)
		}

	}

}

func main() {

	a := flowgraph.MakeConn(false,true,int(0))
	b := flowgraph.MakeConn(false,true,int(0))
	x := flowgraph.MakeConn(false,true,nil)
	g := flowgraph.MakeConn(true,false,nil)

	go tbi(g, a, b)
	go flowgraph.AddFunc(a, b, x)
	go tbo(x, g)

	time.Sleep(1000000000)

}

