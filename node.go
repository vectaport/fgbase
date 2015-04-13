package flowgraph

import (
	"reflect"
	"runtime"
	"sync/atomic"
)

// Node of a flowgraph
type Node struct {
	ID int64                   // unique id
	Name string                // for tracing
	Cnt int64                  // execution count
	Srcs []*Edge               // upstream links
	Dsts []*Edge               // downstream links
	RdyFunc NodeRdy            // func to test Edge readiness
	FireFunc NodeFire          // func to fire Node execution
	Cases []reflect.SelectCase // select cases to read from Edge's
}

// NodeRdy is the function signature for evaluating readiness of Node to fire.
type NodeRdy func(*Node) bool

// NodeFire is the function signature for firing off a flowgraph primitive (or stub).
// Any error message should be written using Node.Errorf and
// nil written to any output Edge.

type NodeFire func(*Node)

// MakeNode returns a new Node with slices of input and output Edge's and functions for testing readiness then firing.
func MakeNode(
	name string, 
	srcs, dsts []*Edge, 
	ready NodeRdy, 
	fire NodeFire) Node {
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

func prefixTracel(n *Node) (format string, tracel []interface {}) {
	var varl [] interface {}
	varl = append(varl, n.Name)
	varl = append(varl, n.ID)
	if (n.Cnt>=0) {
		varl = append(varl, n.Cnt)
	} else {
		varl = append(varl, "*")
	}
	var f string
	if (TraceIndent) {
		for i := int64(0);i<n.ID;i++ {
			f += "\t"
		}
	}
	f += "%s(%d:%v) "
	return f,varl
}

func addSliceToTracel(d Datum, format string, tracel []interface {}) (newfmt string, newtracel []interface {}) {
	m := 8
	l := Len(d)
	if l < m || TraceLevel==VVV { m = l }
	tracel = append(tracel, d)
	format += "%T(["
	for i := 0; i<m; i++ {
		if i!=0 {format += " "}
		tracel = append(tracel, Index(d,i))
		format += "%+v"
	}
	if m<l && TraceLevel<VVV {format += " ..."}
	format += "])"
	return format,tracel
}

// Tracef for debug trace printing.  Uses atomic log mechanism.
func (n *Node) Tracef(format string, v ...interface{}) {
	if (TraceLevel<V) {
		return
	}
	newfmt,tracel := prefixTracel(n)
	newfmt += format
	tracel = append(tracel, v...)
	StdoutLog.Printf(newfmt, tracel...)
}

// Errorf for logging of error messages.  Uses atomic log mechanism.
func (n *Node) Errorf(format string, v ...interface{}) {
	_,nm,ln,_ := runtime.Caller(1)
	newfmt,tracel := prefixTracel(n)
	newfmt += format
	tracel = append(tracel, v...)
	newfmt += " -- %s:%d "
	tracel = append(tracel, nm)
	tracel = append(tracel, ln)
	StderrLog.Printf(newfmt, tracel...)
}

// TraceValRdy lists Node input values and output readiness
func (n *Node) TraceValRdy(valOnly bool) {

	if (!valOnly && TraceLevel<VV || TraceLevel==Q)  {return}
	newfmt,tracel := prefixTracel(n)
	if !valOnly { newfmt += "<<" }
	for i := range n.Srcs {
		srci := n.Srcs[i]
		if (i!=0) { newfmt += "," }
		tracel = append(tracel, srci.Name)
		newfmt += "%s="
		if (srci.Rdy) {
			if IsSlice(srci.Val) {
				newfmt,tracel = addSliceToTracel(srci.Val, newfmt, tracel)
			} else {
				if true { 
					if srci.Val==nil  {
						newfmt += "<nil>"
					} else {
						tracel = append(tracel, srci.Val)
						tracel = append(tracel, srci.Val)
						newfmt += "%T(%v)"
					}
				} else {
					tracel = append(tracel, srci)
					newfmt += "%+v"
				}
			}
		} else {
			newfmt += "{}"
		}
	}
	newfmt += ":"
	for i := range n.Dsts {
		dsti := n.Dsts[i]
		if (i!=0) { newfmt += "," }
		if (valOnly) {
			tracel = append(tracel, dsti.Name)
			newfmt += "%s="
			if IsSlice(dsti.Val) {
				newfmt,tracel = addSliceToTracel(dsti.Val, newfmt, tracel)
			} else {
				if (dsti.Val != nil) {
					tracel = append(tracel, dsti.Val)
					tracel = append(tracel, dsti.Val)
					newfmt += "%T(%v)"
				} else {
					newfmt += func () string { 
						if (dsti.NoOut) { 
							return "{}" 
						}
						return "<nil>" 
					} ()
				}
			}
		} else {
			if true {
				tracel = append(tracel, dsti.Name+".Rdy")
				tracel = append(tracel, dsti.Rdy)
				newfmt += "%s=%v"
			} else {
				tracel = append(tracel, dsti.Name)
				tracel = append(tracel, dsti)
				newfmt += "%s=%+v"
			}
		}
	}
	if !valOnly { newfmt += ">>" }
	newfmt += "\n"
	StdoutLog.Printf(newfmt, tracel...)
}

// TraceVals lists input and output values for a Node.
func (n *Node) TraceVals() { n.TraceValRdy(true) }

// IncrExecCnt increments execution count of Node
func (n *Node) IncrExecCnt() {
	if (GlobalStats) {
		c := atomic.AddInt64(&globalFireCnt, 1)
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
	i,recv,recvOK := reflect.Select(n.Cases)
	if (recvOK) {
		if i<l {
			srci := n.Srcs[i]
			srci.Val = recv.Interface()
			srci.Rdy = true
			if (TraceLevel>=VV) {
				if (srci.Val==nil) {
					n.Tracef("<nil> <- %s.Data\n", srci.Name)
				} else {
					n.Tracef("%T(%v) <- %s.Data\n", srci.Val, srci.Val, srci.Name)
				}
			}
		} else {
			dsti := n.Dsts[i-l]
			dsti.Rdy = true
			if (TraceLevel>=VV) {
				n.Tracef("true <- %s.Ack\n", dsti.Name)
			}
		}
	}
}

// Run is an event loop that runs forever for each Node.
func (n *Node) Run() {
	for {
		if(n.RdyAll()) {
			n.Fire()	
			n.SendAll()
		}
		n.RecvOne()
	}
}

