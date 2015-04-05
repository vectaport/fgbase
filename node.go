package flowgraph

import (
	"reflect"
	"sync/atomic"
)

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

// RdyTest is the function signature for evaluating readiness of Node to fire.
type RdyTest func(*Node) bool

// FireNode is the function signature for firing off flowgraph stub.
type FireNode func(*Node)

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
	if (Indent) {
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
	if l < m { m = l }
	tracel = append(tracel, d)
	format += "%T(["
	for i := 0; i<m; i++ {
		if i!=0 {format += " "}
		tracel = append(tracel, Index(d,i))
		format += "%+v"
	}
	if m<l {format += " ..."}
	format += "])"
	return format,tracel
}

// Tracef for debug trace printing.  Uses atomic log mechanism.
func (n *Node) Tracef(format string, v ...interface{}) {
	if (!Debug) {
		return
	}
	newfmt,tracel := prefixTracel(n)
	newfmt += format
	tracel = append(tracel, v...)
	StdoutLog.Printf(newfmt, tracel...)
}

// TraceValRdy lists Node input values and output readiness
func (n *Node) TraceValRdy(valOnly bool) {

	if (!valOnly && !Debug) {return}
	newfmt,tracel := prefixTracel(n)
	if !valOnly { newfmt += "<<" }
	for i := range n.Srcs {
		if (i!=0) { newfmt += "," }
		tracel = append(tracel, n.Srcs[i].Name)
		newfmt += "%s="
		if (n.Srcs[i].Rdy) {
			if IsSlice(n.Srcs[i].Val) {
				newfmt,tracel = addSliceToTracel(n.Srcs[i].Val, newfmt, tracel)
			} else {
				if true { 
					tracel = append(tracel, n.Srcs[i].Val)
					newfmt += "%v"
				} else {
					tracel = append(tracel, n.Srcs[i])
					newfmt += "%+v"
				}
			}
		} else {
			tracel = append(tracel, "{}")
			newfmt += "%s"
		}
	}
	newfmt += ":"
	for i := range n.Dsts {
		if (i!=0) { newfmt += "," }
		if (valOnly) {
			tracel = append(tracel, n.Dsts[i].Name)
			newfmt += "%s="
			if IsSlice(n.Dsts[i].Val) {
				newfmt,tracel = addSliceToTracel(n.Dsts[i].Val, newfmt, tracel)
			} else {
				if (n.Dsts[i].Val != nil) {
					tracel = append(tracel, n.Dsts[i].Val)
					tracel = append(tracel, n.Dsts[i].Val)
					newfmt += "%T(%v)"
				} else {
					tracel = append(tracel, "{}")
					newfmt += "%v"
				}
			}
		} else {
			if true {
				tracel = append(tracel, n.Dsts[i].Name+".Rdy")
				tracel = append(tracel, n.Dsts[i].Rdy)
				newfmt += "%s=%v"
			} else {
				tracel = append(tracel, n.Dsts[i].Name)
				tracel = append(tracel, n.Dsts[i])
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

