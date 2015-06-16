package flowgraph

import (		
)      			

func fifoRdy(n *Node) bool {

	// ready when channel isn't full and there is input
	a := n.Srcs[0]
	fifo := a.Aux.(chan Datum)
	if a.SrcRdy(n) && cap(fifo)>len(fifo) {
		return true
	} 

	// ready when channel isn't empty and output has been requested.
	x := n.Dsts[0]
	return x.DstRdy(n) && len(fifo)>0
	
}

func fifoFire(n *Node) {

	a := n.Srcs[0]
	x := n.Dsts[0]
	fifo := a.Aux.(chan Datum)

	lenBefore := len(fifo)

	// if input is ready, write to fifo
	if a.SrcRdy(n) && cap(fifo)>len(fifo) {
		fifo <- a.Val
	} else {
		a.NoOut = true
	}

	// if output is ready read from fifo and set up for write to output
	if x.DstRdy(n) && len(fifo)>0 { 
		x.Val = <- fifo
	} else {
		x.NoOut = true
	}

	if TraceLevel>Q { 
		n.Tracef("FIFO SIZE %d --> %d\n", lenBefore, len(fifo))
		if lenBefore==len(fifo) {
			n.Tracef("FIFO STABLE at %d\n", lenBefore)
		}
	}

}

// FuncFIFO uses a channel to implement a buffer or queue.
func FuncFIFO(a, x Edge, sz int) Node {

	node := MakeNode("fifo", []*Edge{&a}, []*Edge{&x}, fifoRdy, fifoFire)
	a.Aux = make(chan Datum, sz)
	return node
}
	
