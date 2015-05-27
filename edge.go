package flowgraph

import (
	"fmt"
	"strconv"
)

type Nada struct {}

// Edge of a flowgraph.
type Edge struct {

	// values shared by upstream and downstream Node
	Name string        // for trace
	Data *[]chan Datum // slice of data channels
	Ack chan Nada      // request (or acknowledge) channel

	// values unique to upstream and downstream Node
	Val Datum          // generic empty interface
	RdyCnt int         // readiness of I/O
	NoOut bool         // set true to inhibit one output, data or ack
	Aux Datum          // auxiliary empty interface to hold state
	Ack2 chan Nada     // alternate channel for ack steering

}

// Return new Edge to connect one upstream Node to one or more downstream Node's.
// Initialize optional data value to start flow.
func makeEdge(name string, initVal Datum) Edge {
	var e Edge
	e.Name = name
	e.Val = initVal
	dc := make([]chan Datum, 0)
	e.Data = &dc
	e.Ack = make(chan Nada, ChannelSize)
	return e
}

// MakeEdge initializes optional data value to start flow.
func MakeEdge(name string, initVal Datum) Edge {
	return makeEdge(name, initVal)
}

// Const sets up an Edge to provide a constant value.
func (e *Edge) Const(d Datum) {
	e.Val = d
	e.Data = nil
	e.Ack = nil
}
	
// IsConst returns true if Edge provides a constant value.
func (e *Edge) IsConst() bool { 
	return e.Data == nil && e.Val != nil
}

// Sink sets up an Edge as a value sink.
func (e *Edge) Sink() {
	e.Val = nil
	e.Data = nil
	e.Ack = nil
}

// IsSink returns true if Edge is a value sink.
func (e *Edge) IsSink() bool { 
	return e.Data == nil && e.Val == nil
}

// Rdy tests if RdyCnt has returned to zero.
func (e *Edge) Rdy() bool {
	return e.RdyCnt==0
}

// SrcReadRdy tests if a source edge is ready for a data read.
func (e *Edge) SrcReadRdy() bool {
	return len((*e.Data)[0])>0
}

// SrcWriteRdy tests if a source edge is ready for an ack write.
func (e *Edge) SrcWriteRdy() bool {
	return len(e.Ack)<cap(e.Ack)
}

// DstReadRdy tests if a destination edge is ready for an ack read.
func (e *Edge) DstReadRdy() bool {
	return len(e.Ack)>0
}

// DstWriteRdy tests if a destination edge is ready for a data write.
func (e *Edge) DstWriteRdy() bool {
	for _,c := range *e.Data {
		if cap(c)==len(c) { return false }
	}
	return true
}

// SendData writes to the Data channel
func (e *Edge) SendData(n *Node) {
	if(e.Data !=nil) {
		if (!e.NoOut) {
			if (TraceLevel>=VV) {
				nm := e.Name + ".Data"
				if len(*e.Data)>1 {
					nm += "{" + strconv.Itoa(len(*e.Data)) + "}"
				}
				ev := e.Val
				var asterisk string
				
				// remove from wrapper if in one
				if _,ok := ev.(nodeWrap); ok {
					n2 := ev.(nodeWrap).node
					ev = ev.(nodeWrap).datum
					asterisk += fmt.Sprintf(" *(Ack2=%p)", n2.Srcs[0].Ack)
				}

				if (ev==nil) {
					n.Tracef("%s <- <nil>%s\n", nm, asterisk)
				} else {
					n.Tracef("%s <- %s%s\n", nm, String(ev), asterisk)
				}
			}

			for i := range *e.Data {
				(*e.Data)[i] <- e.Val
				// n.Tracef("wrote data cap=%d, len=%d\n", cap((*e.Data)[i]), len((*e.Data)[i]))
			}
			e.RdyCnt += len(*e.Data)
			e.Val = nil
		} else {
			e.NoOut = false
		}
	}
}

// SendAck writes Nada to the Ack channel
func (e *Edge) SendAck(n *Node) {
	if(e.Ack !=nil) {
		if (!e.NoOut) {
			var nada Nada
			if e.Ack2 != nil {
				if (TraceLevel>=VV) {
					n.Tracef("%s.Ack2(%p) <-\n", e.Name, e.Ack2)
				}
				e.Ack2 <- nada
				e.Ack2 = nil
			} else {
				if (TraceLevel>=VV) {
					n.Tracef("%s.Ack <-\n", e.Name)
				}
				e.Ack <- nada
				// n.Tracef("wrote ack cap=%d, len=%d\n", cap(e.Ack), len(e.Ack))
			}
			e.RdyCnt = 1
			
		} else {
			e.NoOut = false
		}
	}
}

// MakeEdges returns a slice of Edge.
func MakeEdges(sz int) []Edge {
	e := make([]Edge, sz)
	for i:=0; i<sz; i++ {
		nm := "e" + strconv.Itoa(int(i))
		e[i] = MakeEdge(nm, nil)
	}
	return e
}

// PoolEdge returns an output Edge that is directed back into the Pool.
func (dst *Edge) PoolEdge(src *Edge) Edge {
	e := *dst
	e.Data = src.Data
	e.Name = dst.Name+"("+src.Name+")"
	return e
}
	
