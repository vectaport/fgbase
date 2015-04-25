package flowgraph

import (
	"reflect"
	"runtime"
	"strconv"
	"sync/atomic"
)

type edgeDir struct {
	edge *Edge
	srcFlag bool
}
	

// Node of a flowgraph.
type Node struct {
	ID int64                        // unique id
	Name string                     // for tracing
	Cnt int64                       // execution count
	Srcs []*Edge                    // upstream links
	Dsts []*Edge                    // downstream links
	RdyFunc NodeRdy                 // func to test Edge readiness
	FireFunc NodeFire               // func to trigger Node execution
	RunFunc NodeRun                 // func to repeatedly run Node execution

	cases []reflect.SelectCase      // select cases to read from Edge's
	caseToEdgeDir map [int] edgeDir // map from selected case to associated Edge
}

// NodeRdy is the function signature for evaluating readiness of a Node to execute.
type NodeRdy func(*Node) bool

// NodeFire is the function signature for executing a Node.
// Any error message should be written using Node.Errorf and
// nil written to any output Edge.
type NodeFire func(*Node)

// NodeRun is the function signature for an alternate Node event loop.
type NodeRun func(*Node)

// MakeNode returns a new Node with slices of input and output Edge's and functions for testing readiness then firing.
func MakeNode(
	name string, 
	srcs, dsts []*Edge, 
	ready NodeRdy, 
	fire NodeFire) Node {
	var n Node
	i := atomic.AddInt64(&NodeID, 1)
	n.ID = i-1
	n.Name = name
	n.Cnt = -1
	n.Srcs = srcs
	n.Dsts = dsts
	n.RdyFunc = ready
	n.FireFunc = fire
	n.caseToEdgeDir = make(map[int]edgeDir)
	var cnt = 0
	for i := range n.Srcs {
		n.Srcs[i].RdyCnt = func () int {
			if n.Srcs[i].Val!=nil { return 0 }; return 1}()
		if n.Srcs[i].Data != nil {
			j := len(*n.Srcs[i].Data)
			*n.Srcs[i].Data = append(*n.Srcs[i].Data, make(chan Datum))
			n.cases = append(n.cases, reflect.SelectCase{Dir:reflect.SelectRecv, Chan:reflect.ValueOf((*n.Srcs[i].Data)[j])})
			n.caseToEdgeDir[cnt] = edgeDir{n.Srcs[i], true}
			cnt = cnt+1
		}
	}
	for i := range n.Dsts {
		n.Dsts[i].RdyCnt = func (b bool) int {if b { return 0 }; return len(*n.Dsts[i].Data) } (n.Dsts[i].Val==nil)
		if n.Dsts[i].Ack!=nil {
			n.cases = append(n.cases, reflect.SelectCase{Dir:reflect.SelectRecv, Chan:reflect.ValueOf(n.Dsts[i].Ack)})
			n.caseToEdgeDir[cnt] = edgeDir{n.Dsts[i], false}
			cnt = cnt+1
		}
	}

	return n
}

func prefixTracel(n *Node) (format string, tracel []interface {}) {
	var addNodeAddr = false
	var varl [] interface {}
	varl = append(varl, n.Name)
	varl = append(varl, n.ID)
	if (n.Cnt>=0) {
		varl = append(varl, n.Cnt)
	} else {
		varl = append(varl, "*")
	}
	if (addNodeAddr) { varl = append(varl, n) }
	var f string
	if (TraceIndent) {
		for i := int64(0);i<n.ID;i++ {
			f += "\t"
		}
	}
	if (addNodeAddr) { f += "%s(%d:%v:%p) " } else { f += "%s(%d:%v) " }
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

// TraceValRdy lists Node input values and output values or readiness.
func (n *Node) TraceValRdy(valOnly bool) {

	if (!valOnly && TraceLevel<VVV || TraceLevel==Q)  {return}
	newfmt,tracel := prefixTracel(n)
	if !valOnly { newfmt += "<<" }
	for i := range n.Srcs {
		srci := n.Srcs[i]
		if (i!=0) { newfmt += "," }
		tracel = append(tracel, srci.Name)
		newfmt += "%s"
		if (srci.Rdy()) {
			if IsSlice(srci.Val) {
				newfmt,tracel = addSliceToTracel(srci.Val, newfmt, tracel)
			} else {
				if srci.Val==nil  {
					newfmt += "=<nil>"
				} else {
					tracel = append(tracel, srci.Val)
					tracel = append(tracel, srci.Val)
					newfmt += "=%T(%v)"
				}
			}
		} else {
			newfmt += "={}"
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
				tracel = append(tracel, dsti.Name)
				if dsti.RdyCnt==1 {
					newfmt += "%s={}"
				} else {
					tracel = append(tracel, dsti.RdyCnt)
					newfmt += "%s={%v}"
				}
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

// IncrFireCnt increments execution count of Node.
func (n *Node) IncrFireCnt() {
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
			if !n.Srcs[i].Rdy() { return false }
		}
		for i := range n.Dsts {
			if !n.Dsts[i].Rdy() { return false }
		}
	} else {
		if !n.RdyFunc(n) { return false }
	}
	n.IncrFireCnt();
	return true
}

// Fire node using function pointer.
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

// RecvOne reads one data or ack and decrements RdyCnt.
func (n *Node) RecvOne() {
	n.TraceValRdy(false)
	i,recv,recvOK := reflect.Select(n.cases)
	if (recvOK) {
		if n.caseToEdgeDir[i].srcFlag {
			srci := n.caseToEdgeDir[i].edge
			srci.Val = recv.Interface()
			srci.RdyCnt--
			if (TraceLevel>=VV) {
				if (srci.Val==nil) {
					n.Tracef("<nil> <- %s.Data\n", srci.Name)
				} else {
					n.Tracef("%T(%v) <- %s.Data\n", srci.Val, srci.Val, srci.Name)
				}
			}
		} else {
			dsti := n.caseToEdgeDir[i].edge
			dsti.RdyCnt--
			if (TraceLevel>=VV) {
				nm := dsti.Name + ".Ack"
				if len(*dsti.Data)>1 {
					nm += "{" + strconv.Itoa(dsti.RdyCnt+1) + "}"
				}
				n.Tracef("<- %s\n", nm)
			}
		}
	}
}

// Run is an event loop that runs forever for each Node.
func (n *Node) Run() {
	if n.RunFunc != nil {
		n.RunFunc(n)
		return
	}

	for {
		if n.RdyAll() {
			n.Fire()	
			n.SendAll()
		}
		n.RecvOne()
	}
}

// MakeNodes returns a slice of Node.
func MakeNodes(sz int) []Node {
	n := make([]Node, sz)
	return n
}

