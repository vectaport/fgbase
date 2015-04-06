package flowgraph

// Edge of a flowgraph.
type Edge struct {

	// values shared by upstream and downstream Node
	Name string       // for trace
	Data chan Datum   // downstream data channel
	Ack chan bool     // upstream request channel

	// values unique to upstream and downstream Node
	Val Datum         // generic empty interface
	Rdy bool          // readiness of I/O
	Nack bool         // set true to inhibit acking
	Aux Datum         // auxiliary empty interface to hold state
}

// Return new Edge to connect two Node's.
// Initialize optional data value to start flow.
func newEdge(name string, initVal Datum, data chan Datum, ack chan bool) Edge {
	var e Edge
	e.Name = name
	e.Val = initVal
	e.Data = data
	e.Ack = ack
	return e
}

// MakeEdge initializes optional data value to start flow.
func MakeEdge(name string, initVal Datum) Edge {
	return newEdge(name, initVal, make(chan Datum), make(chan bool))
}

// MakeEdgeConst initializes a dangling edge to provide a constant value.
func MakeEdgeConst(name string, initVal Datum) Edge {
	return newEdge(name, initVal, nil, nil)
}

// MakeEdgeSink initializes a dangling edge to provide a sink for values.
func MakeEdgeSink(name string) Edge {
	return newEdge(name, nil, nil, nil)
}
// IsConstant returns true if Edge is an implied constant
func (e *Edge) IsConstant() bool { 
	return e.Ack == nil && e.Val != nil
}

// IsSink returns true if Edge is an implied sink
func (e *Edge) IsSink() bool { 
	return e.Ack == nil && e.Val == nil
}

// SendData writes to the Data channel
func (e *Edge) SendData(n *Node) {
	if(e.Data !=nil && e.Val != nil) {
		if (TraceLevel>=VV) {
			n.Tracef("%s.Data <- %v\n", e.Name, e.Val)
		}
		e.Data <- e.Val
		e.Rdy = false
		e.Val = nil
	}
}

// SendAck writes true to the Ack channel
func (e *Edge) SendAck(n *Node) {
	if(e.Ack !=nil) {
		if (!e.Nack) {
			if (TraceLevel>=VV) {
				n.Tracef("%s.Ack <- true\n", e.Name)
			}
			e.Ack <- true
			e.Rdy = false
		} else {
			e.Nack = false
		}
	}
}


