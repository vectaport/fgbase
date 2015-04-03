package flowgraph

import (
	"sync/atomic"
	"fmt"
	"reflect"
)

var node_id int64 = 0
var global_exec_cnt int64 = 0

// Enable debug tracing
var Debug bool = false

// Indent trace by node id
var Indent bool = false

// Use global execution count
var GlobalExecCnt bool = false


// Datum is an empty interface for generic data flow.
type Datum interface{}

// RdyTest is the function signature for evaluating readiness of Node to fire.
type RdyTest func(*Node) bool

// Firenode is the function signature for firing off flowgraph stub.
type FireNode func(*Node)

// Edge of a flowgraph.
type Edge struct {

	// values shared by upstream and downstream Node
	Name string       // for trace
	Data chan Datum   // downstream data channel
	Ack chan bool     // upstream request channel

	// values unique to upstream and downstream Node
	Val Datum         // generic data
	Rdy bool          // readiness of I/O
	Nack bool         // set true to inhibit acking
	Aux Datum         // auxiliary empty interface to hold state
}

// Return new Edge to connect two Node's.
// Initialize optional data value to start flow.
func new_edge(name string, init_val Datum, data chan Datum, ack chan bool) Edge {
	var e Edge
	e.Name = name
	e.Val = init_val
	e.Data = data
	e.Ack = ack
	return e
}

// MakeEdge initializes optional data value to start flow.
func MakeEdge(name string, init_val Datum) Edge {
	return new_edge(name, init_val, make(chan Datum), make(chan bool))
}

// MakeEdgeConst initializes a dangling edge to provide a constant value.
func MakeEdgeConst(name string, init_val Datum) Edge {
	return new_edge(name, init_val, nil, nil)
}

// MakeEdgeSink initializes a dangling edge to provide a sink for values.
func MakeEdgeSink(name string) Edge {
	return new_edge(name, nil, nil, nil)
}
// IsConstat returns true if Edge is an implied constant
func IsConstant(e *Edge) bool { 
	return e.Ack == nil && e.Val != nil
}

// IsSink returns true if Edge is an implied sink
func IsSink(e *Edge) bool { 
	return e.Ack == nil && e.Val == nil
}

// SendData writes to the Data channel
func (e *Edge) SendData(n *Node) {
	if(e.Data !=nil && e.Val != nil) {
		n.Tracef("%s.Data <- %v\n", e.Name, e.Val)
		e.Data <- e.Val
		e.Rdy = false
		e.Val = nil
	}
}

// SendAck writes true to the Ack channel
func (e *Edge) SendAck(n *Node) {
	if(e.Ack !=nil) {
		if (!e.Nack) {
			n.Tracef("%s.Ack <- true\n", e.Name)
			e.Ack <- true
			e.Rdy = false
		} else {
			e.Nack = false
		}
	}
}

// flowgraph Node
type Node struct {
	Id int64                   // unique id
	Name string                // for tracing
	Cnt int64                  // execution count
	Srcs []*Edge               // upstream links
	Dsts []*Edge               // downstream links
	RdyFunc RdyTest            // func to test Edge readiness
	FireFunc FireNode          // func to fire Node execution
	Cases []reflect.SelectCase // select cases to read from Edge's
}

// Return new Node with slices of input and output Edge's and customizable ready-testing function
func MakeNode(name string, srcs, dsts []*Edge, ready RdyTest, fire FireNode) Node {
	var n Node
	i := atomic.AddInt64(&node_id, 1)
	n.Id = i-1
	n.Name = name
	n.Cnt = -1
	n.Srcs = srcs
	n.Dsts = dsts
	var casel [] reflect.SelectCase
	for i := range n.Srcs {
		n.Srcs[i].Rdy = n.Srcs[i].Val!=nil
		casel = append(casel, reflect.SelectCase{Dir:reflect.SelectRecv, Chan:reflect.ValueOf(n.Srcs[i].Data)})
	}
	for i := range n.Dsts {
		n.Dsts[i].Rdy = n.Dsts[i].Val==nil
		casel = append(casel, reflect.SelectCase{Dir:reflect.SelectRecv, Chan:reflect.ValueOf(n.Dsts[i].Ack)})
	}
	n.Cases = casel
	n.RdyFunc = ready
	n.FireFunc = fire
	return n
}

func prefix_varlist(n *Node) (format string, varlist []interface {}) {
	var varl [] interface {}
	varl = append(varl, n.Name)
	varl = append(varl, n.Id)
	if (n.Cnt>=0) {
		varl = append(varl, n.Cnt)
	} else {
		varl = append(varl, "*")
	}
	var f string
	if (Indent) {
		for i := int64(0);i<n.Id;i++ {
			f += "\t"
		}
	}
	f += "%s(%d:%v) "
	return f,varl
}

// Debug trace printing
func (n *Node) Tracef(format string, v ...interface{}) {
	if (!Debug /*|| format=="select\n"*/) {
		return
	}
	newfmt,varlist := prefix_varlist(n)
	newfmt += format
	varlist = append(varlist, v...)
	fmt.Printf(newfmt, varlist...)
}

// Trace Node input values and output readiness
func (n *Node) TraceValRdy(val_only bool) {
	if (!val_only && !Debug) {return}
	newfmt,varlist := prefix_varlist(n)
	if !val_only { newfmt += "[" }
	for i := range n.Srcs {
		if (i!=0) { newfmt += "," }
		varlist = append(varlist, n.Srcs[i].Name)
		newfmt += "%s="
		if (n.Srcs[i].Rdy) {
			varlist = append(varlist, n.Srcs[i].Val)
			varlist = append(varlist, n.Srcs[i].Val)
			newfmt += "%T(%v)"
//			varlist = append(varlist, n.Srcs[i])
//			newfmt += "%+v"
		} else {
			varlist = append(varlist, "{}")
			newfmt += "%s"
		}
	}
	newfmt += ":"
	for i := range n.Dsts {
		if (i!=0) { newfmt += "," }
		if (val_only) {
			varlist = append(varlist, n.Dsts[i].Name)
			newfmt += "%s="
			if (n.Dsts[i].Val != nil) {
				varlist = append(varlist, n.Dsts[i].Val)
				varlist = append(varlist, n.Dsts[i].Val)
				newfmt += "%T(%v)"
			} else {
				varlist = append(varlist, "{}")
				newfmt += "%v"
			}
		} else {
			varlist = append(varlist, n.Dsts[i].Name+".Rdy")
			varlist = append(varlist, n.Dsts[i].Rdy)
			newfmt += "%s=%v"
//			varlist = append(varlist, n.Dsts[i].Name)
//			varlist = append(varlist, n.Dsts[i])
//			newfmt += "%s=%+v"
		}
	}
	if !val_only { newfmt += "]" }
	newfmt += "\n"
	fmt.Printf(newfmt, varlist...)
}

// Tracing Node execution
func (n *Node) TraceVals() { n.TraceValRdy(true) }

// Increment execution count of Node
func (n *Node) IncrExecCnt() {
	if (GlobalExecCnt) {
		c := atomic.AddInt64(&global_exec_cnt, 1)
		n.Cnt = c-1
	} else {
		n.Cnt = n.Cnt+1
	}
}

// Test readiness of Node to execute
func (n *Node) RdyAll() bool {
	if (n.RdyFunc == nil) {
		for i := range n.Srcs {
			if !n.Srcs[i].Rdy { return false }
		}
		for i := range n.Dsts {
			if !n.Dsts[i].Rdy { return false }
		}
	} else {
		if !n.RdyFunc(n) { return false }
	}
	n.IncrExecCnt();
	return true
}

// Fire node using function pointer
func (n *Node) Fire() {
	if (n.FireFunc!=nil) { n.FireFunc(n) }
}


// Sink value (to avoid unused error)
func Sink(a Datum) () {
}

// Test value for zero
func ZeroTest(a Datum) bool {

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
	default: { return false }
	}
}

// Send all data and acks after new result is computed
func (n *Node) SendAll() {
	n.TraceVals()
	for i := range n.Srcs {
		n.Srcs[i].SendAck(n)
	}
	for i := range n.Dsts {
		n.Dsts[i].SendData(n)
	}
}

// Receive one data or ack and mark that input as ready
func (n *Node) RecvOne() {
	l := len(n.Srcs)
	n.TraceValRdy(false)
	chosen,recv,recvOK := reflect.Select(n.Cases)
	if (recvOK) {
		if chosen<l {
			n.Srcs[chosen].Val = recv.Interface()
			n.Srcs[chosen].Rdy = true
			n.Tracef("%T(%v) <- %s.Data\n", n.Srcs[chosen].Val, n.Srcs[chosen].Val, n.Srcs[chosen].Name)
		} else {
			n.Dsts[chosen-l].Rdy = true
			n.Tracef("true <- %s.Ack\n", n.Dsts[chosen-l].Name)
		}
	}
}

// Event loop to run forever
func (n *Node) Run() {
	for {
		if(n.RdyAll()) {
			n.Fire()
			n.SendAll()
		}

		n.RecvOne()
	}
}

