package flowgraph

import (
	"math"
	"testing"
	"time"
)

func tbiAdd(x, y Edge) Node {

	node := MakeNode("tbi", nil, []*Edge{&x, &y}, nil, nil)
	node.RunFunc = tbiAddRun
	return node
}

func tbiAddRun (node *Node) {
	x := node.Dsts[0]
	y := node.Dsts[1]

	x.Aux = 0
	y.Aux = 0
	var i int = 0
	for {
		if (i>10) { break }
		if node.RdyAll(){
			x.Val = x.Aux
			y.Val = y.Aux
			x.Aux = x.Aux.(int) + 1
			y.Aux = y.Aux.(int) + 1
			node.TraceVals()
			node.SendAll()
			i = i + 1
		}
		node.RecvOne()
	}

	x.Aux = float32(0)
	y.Aux = float32(0)
	i = 0
	for {
		if (i>9) { break }
		if node.RdyAll(){
			x.Val = x.Aux
			y.Val = y.Aux
			x.Aux = x.Aux.(float32) + 1
			y.Aux = y.Aux.(float32) + 1
			node.TraceVals()
			node.SendAll()
			i = i + 1
		}
		node.RecvOne()
	}

	x.Aux = uint64(math.MaxUint64)
	y.Aux = -1
	i = 0
	for {
		if (i > 0) { break }
		if node.RdyAll(){
			x.Val = x.Aux
			y.Val = y.Aux
			node.TraceVals()
			node.SendAll()
			i = i + 1
		}
		node.RecvOne()
	}

	x.Aux = int8(0)
	y.Aux = uint64(0)
	i = 0
	for  {
		if (i > 0) { break }
		if node.RdyAll() {
			x.Val = x.Aux
			y.Val = y.Aux
			node.TraceVals()
			node.SendAll()
			i = i + 1
		}
		node.RecvOne()
	}

	x.Aux = int8(0)
	y.Aux = int16(0)
	i = 0
	for  {
		if (i > 0) { break }
		if node.RdyAll() {
			x.Val = x.Aux
			y.Val = y.Aux
			node.TraceVals()
			node.SendAll()
			i = i + 1
		}
		node.RecvOne()
	}


	x.Aux = "Can you add an int to a string?"
	y.Aux = int8(77)
	i = 0
	for  {
		if (i > 0) { break }
		if node.RdyAll() {
			x.Val = x.Aux
			y.Val = y.Aux
			node.TraceVals()
			node.SendAll()
			i = i + 1
		}
		node.RecvOne()
	}

	var arr = [4]complex128 {0+0i,0+0i,0+0i,0+0i}
	x.Aux = arr[:]
	y.Aux = int8(77)
	i = 0
	for  {
		if (i > 0) { break }
		if node.RdyAll() {
			x.Val = x.Aux
			y.Val = y.Aux
			node.TraceVals()
			node.SendAll()
			i = i + 1
		}
		node.RecvOne()
	}
	

	// read all the acks to clean up
	for  {
		node.RecvOne()
	}
	

}

func tboAdd(a Edge) Node {

	node := MakeNode("tbo", []*Edge{&a}, nil, nil, nil)
	return node

}

func TestAdd(t *testing.T) {

	TraceLevel = V

	e,n := MakeGraph(3,3)

	n[0] = tbiAdd(e[0], e[1])
	n[1] = FuncAdd(e[0], e[1], e[2])
	n[2] = tboAdd(e[2])

	RunAll(n, time.Second)

}

