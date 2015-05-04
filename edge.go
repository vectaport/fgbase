package flowgraph

import (
	"fmt"
	"strconv"
)

// Edge of a flowgraph.
type Edge struct {

	// values shared by upstream and downstream Node
	Name string        // for trace
	Data *[]chan Datum // slice of data channels
	Ack chan bool      // request (or acknowledge) channel

	// values unique to upstream and downstream Node
	Val Datum          // generic empty interface
	RdyCnt int         // readiness of I/O
	NoOut bool         // set true to inhibit one output, data or ack
	Aux Datum          // auxiliary empty interface to hold state
	Ack2 chan bool     // alternate channel for ack steering

}

// Return new Edge to connect one upstream Node to one or more downstream Node's.
// Initialize optional data value to start flow.
func newEdge(name string, initVal Datum) Edge {
	var e Edge
	e.Name = name
	e.Val = initVal
	dc := make([]chan Datum, 0)
	e.Data = &dc
	e.Ack = make(chan bool, 1)
	return e
}

// MakeEdge initializes optional data value to start flow.
func MakeEdge(name string, initVal Datum) Edge {
	return newEdge(name, initVal)
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

// Rdy tests if RdyCnt has return to zero.
func (e *Edge) Rdy() bool {
	return e.RdyCnt==0
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
				if _,ok := ev.(ackWrap); ok {
					asterisk += fmt.Sprintf(" *(Ack2=%p)", ev.(ackWrap).ack)
					ev = ev.(ackWrap).d
				}
				if (ev==nil) {
					n.Tracef("%s <- <nil>%s\n", nm, asterisk)
				} else {
					n.Tracef("%s <- %s%s\n", nm, 
						func() string {
							if IsSlice(ev) { return StringSlice(ev) }
							return fmt.Sprintf("%T(%v)", ev, ev)}(),
						asterisk)
				}
			}
			for i := range *e.Data {
				(*e.Data)[i] <- e.Val
			}
			e.RdyCnt = len(*e.Data)
			e.Val = nil
		} else {
			e.NoOut = false
		}
	}
}

// SendAck writes true to the Ack channel
func (e *Edge) SendAck(n *Node) {
	if(e.Ack !=nil) {
		if (!e.NoOut) {
			if e.Ack2 != nil {
				if (TraceLevel>=VV) {
					n.Tracef("%s.Ack2(%p) <-\n", e.Name, e.Ack2)
				}
				e.Ack2 <- true
				e.Ack2 = nil
			} else {
				if (TraceLevel>=VV) {
					n.Tracef("%s.Ack(%p) <-\n", e.Name, e.Ack)
				}
				e.Ack <- true
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
		nm := "e" + strconv.Itoa(i)
		e[i] = MakeEdge(nm, nil)
	}
	return e
}


// AckWrap bundles an ack channel and a Datum to pass acks downstream.
func (e *Edge) AckWrap(d Datum) Datum {
	return ackWrap{e.Ack, d}
}
