/*
Package flowgraph layers a ready-send flow mechanism on top of goroutines.
*/

package flowgraph

import (
	"log"
	"os"
	"reflect"
	"sync/atomic"
)

var nodeID int64
var globalExecCnt int64

// Log for tracing flowgraph execution
var StdoutLog = log.New(os.Stdout, "", 0)

// Enable debug tracing
var Debug = false

// Indent trace by node id
var Indent = false

// Use global execution count
var GlobalExecCnt = false

// RdyTest is the function signature for evaluating readiness of Node to fire.
type RdyTest func(*Node) bool

// FireNode is the function signature for firing off flowgraph stub.
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

// Node of a flowgraph
type Node struct {
	ID int64                   // unique id
	Name string                // for tracing
	Cnt int64                  // execution count
	Srcs []*Edge               // upstream links
	Dsts []*Edge               // downstream links
	RdyFunc RdyTest            // func to test Edge readiness
	FireFunc FireNode          // func to fire Node execution
	Cases []reflect.SelectCase // select cases to read from Edge's
}

// MakeNode returns a new Node with slices of input and output Edge's and functions for testing readiness then firing.
func MakeNode(
	name string, 
	srcs, dsts []*Edge, 
	ready RdyTest, 
	fire FireNode) Node {
	var n Node
	i := atomic.AddInt64(&nodeID, 1)
	n.ID = i-1
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

func prefixVarlist(n *Node) (format string, varlist []interface {}) {
	var varl [] interface {}
	varl = append(varl, n.Name)
	varl = append(varl, n.ID)
	if (n.Cnt>=0) {
		varl = append(varl, n.Cnt)
	} else {
		varl = append(varl, "*")
	}
	var f string
	if (Indent) {
		for i := int64(0);i<n.ID;i++ {
			f += "\t"
		}
	}
	f += "%s(%d:%v) "
	return f,varl
}

func addSliceToVarlist(d Datum, format string, varlist []interface {}) (newfmt string, newvarlist []interface {}) {
	m := 8
	l := Len(d)
	if l < m { m = l }
	varlist = append(varlist, d)
	format += "%T(["
	for i := 0; i<m; i++ {
		if i!=0 {format += " "}
		varlist = append(varlist, Index(d,i))
		format += "%+v"
	}
	if m<l {format += " ..."}
	format += "])"
	return format,varlist
}

// Tracef for debug trace printing.  Uses atomic log mechanism.
func (n *Node) Tracef(format string, v ...interface{}) {
	if (!Debug) {
		return
	}
	newfmt,varlist := prefixVarlist(n)
	newfmt += format
	varlist = append(varlist, v...)
	StdoutLog.Printf(newfmt, varlist...)
}

// TraceValRdy lists Node input values and output readiness
func (n *Node) TraceValRdy(valOnly bool) {

	if (!valOnly && !Debug) {return}
	newfmt,varlist := prefixVarlist(n)
	if !valOnly { newfmt += "<<" }
	for i := range n.Srcs {
		if (i!=0) { newfmt += "," }
		varlist = append(varlist, n.Srcs[i].Name)
		newfmt += "%s="
		if (n.Srcs[i].Rdy) {
			if IsSlice(n.Srcs[i].Val) {
				newfmt,varlist = addSliceToVarlist(n.Srcs[i].Val, newfmt, varlist)
			} else {
				if true { 
					varlist = append(varlist, n.Srcs[i].Val)
					newfmt += "%v"
				} else {
					varlist = append(varlist, n.Srcs[i])
					newfmt += "%+v"
				}
			}
		} else {
			varlist = append(varlist, "{}")
			newfmt += "%s"
		}
	}
	newfmt += ":"
	for i := range n.Dsts {
		if (i!=0) { newfmt += "," }
		if (valOnly) {
			varlist = append(varlist, n.Dsts[i].Name)
			newfmt += "%s="
			if IsSlice(n.Dsts[i].Val) {
				newfmt,varlist = addSliceToVarlist(n.Dsts[i].Val, newfmt, varlist)
			} else {
				if (n.Dsts[i].Val != nil) {
					varlist = append(varlist, n.Dsts[i].Val)
					varlist = append(varlist, n.Dsts[i].Val)
					newfmt += "%T(%v)"
				} else {
					varlist = append(varlist, "{}")
					newfmt += "%v"
				}
			}
		} else {
			if true {
				varlist = append(varlist, n.Dsts[i].Name+".Rdy")
				varlist = append(varlist, n.Dsts[i].Rdy)
				newfmt += "%s=%v"
			} else {
				varlist = append(varlist, n.Dsts[i].Name)
				varlist = append(varlist, n.Dsts[i])
				newfmt += "%s=%+v"
			}
		}
	}
	if !valOnly { newfmt += ">>" }
	newfmt += "\n"
	StdoutLog.Printf(newfmt, varlist...)
}

// TraceVals lists input and output values for a Node.
func (n *Node) TraceVals() { n.TraceValRdy(true) }

// IncrExecCnt increments execution count of Node
func (n *Node) IncrExecCnt() {
	if (GlobalExecCnt) {
		c := atomic.AddInt64(&globalExecCnt, 1)
		n.Cnt = c-1
	} else {
		n.Cnt = n.Cnt+1
	}
}

// RdyAll tests readiness of Node to execute.
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


// SendAll writes all data and acks after new result is computed.
func (n *Node) SendAll() {
	n.TraceVals()
	for i := range n.Srcs {
		n.Srcs[i].SendAck(n)
	}
	for i := range n.Dsts {
		n.Dsts[i].SendData(n)
	}
}

// RecvOne reads one data or ack and marks that input as ready.
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

// Run is an event loop that runs forever.
func (n *Node) Run() {
	for {
		if(n.RdyAll()) {
			n.Fire()	
		n.SendAll()
		}

		n.RecvOne()
	}
}

