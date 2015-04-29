package flowgraph

import (
	"fmt"
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
	WorkFunc NodeWork               // func to do work of the Node
	RunFunc NodeRun                 // func to repeatedly run Node

	cases []reflect.SelectCase      // select cases to read from Edge's
	caseToEdgeDir map [int] edgeDir // map from selected case to associated Edge
}

// NodeRdy is the function signature for evaluating readiness of a Node to execute.
type NodeRdy func(*Node) bool

// NodeWork is the function signature for executing a Node.
// Any error message should be written using Node.Errorf and
// nil written to any output Edge.
type NodeWork func(*Node)

// NodeRun is the function signature for an alternate Node event loop.
type NodeRun func(*Node)

// MakeNode returns a new Node with slices of input and output Edge's and functions for testing readiness then firing.
func MakeNode(
	name string, 
	srcs, dsts []*Edge, 
	ready NodeRdy, 
	work NodeWork) Node {
	var n Node
	i := atomic.AddInt64(&NodeID, 1)
	n.ID = i-1
	n.Name = name
	n.Cnt = -1
	n.Srcs = srcs
	n.Dsts = dsts
	n.RdyFunc = ready
	n.WorkFunc = work
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

func prefixTracef(n *Node) (format string) {
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
	return fmt.Sprintf(f, varl...)
}

func addSliceToTracel(d Datum) string {
	m := 8
	l := Len(d)
	if l < m || TraceLevel==VVV { m = l }
	newFmt := fmt.Sprintf("%T([", d)
	for i := 0; i<m; i++ {
		if i!=0 {newFmt += " "}
		newFmt += fmt.Sprintf("%+v", Index(d,i))
	}
	if m<l && TraceLevel<VVV {newFmt += " ..."}
	newFmt += "])"
	return newFmt
}

// Tracef for debug trace printing.  Uses atomic log mechanism.
func (n *Node) Tracef(format string, v ...interface{}) {
	if (TraceLevel<V) {
		return
	}
	newFmt := prefixTracef(n)
	newFmt += format
	StdoutLog.Printf(newFmt, v...)
}

// Errorf for logging of error messages.  Uses atomic log mechanism.
func (n *Node) Errorf(format string, v ...interface{}) {
	_,nm,ln,_ := runtime.Caller(1)
	newFmt := prefixTracef(n)
	newFmt += format
	newFmt += fmt.Sprintf(" -- %s:%d ", nm, ln)
	StderrLog.Printf(newFmt, v...)
}

// traceValRdySrc lists Node input values
func (n *Node) traceValRdySrc(valOnly bool) string {
	newFmt := prefixTracef(n)
	if !valOnly { newFmt += "<<" }
	for i := range n.Srcs {
		srci := n.Srcs[i]
		if (i!=0) { newFmt += "," }
		newFmt += fmt.Sprintf("%s=", srci.Name)
		if (srci.Rdy()) {
			if IsSlice(srci.Val) {
				newFmt +=addSliceToTracel(srci.Val)
			} else {
				if srci.Val==nil  {
					newFmt += "<nil>"
				} else {
					newFmt += fmt.Sprintf("%T(%v)", srci.Val, srci.Val)
				}
			}
		} else {
			newFmt += "{}"
		}
	}
	newFmt += ":"
	return newFmt
}

// traceValRdyDst lists Node output values or readiness.
func (n *Node) traceValRdyDst(valOnly bool) string {
	var newFmt string
	for i := range n.Dsts {
		dsti := n.Dsts[i]
		if (i!=0) { newFmt += "," }
		if (valOnly) {
			newFmt += fmt.Sprintf("%s=", dsti.Name)
			if IsSlice(dsti.Val) {
				newFmt += addSliceToTracel(dsti.Val)
			} else {
				if (dsti.Val != nil) {
					newFmt += fmt.Sprintf("%T(%v)", dsti.Val, dsti.Val)
				} else {
					newFmt += func () string { 
						if (dsti.NoOut) { 
							return "{}" 
						}
						return "<nil>" 
					} ()
				}
			}
		} else {
			if true {
				if dsti.RdyCnt==1 {
					newFmt += fmt.Sprintf("%s={}", dsti.Name)
				} else {
					newFmt += fmt.Sprintf("%s={%v}", dsti.Name, dsti.RdyCnt)
				}
			} else {
				newFmt += fmt.Sprintf("%s=%+v", dsti.Name, dsti)
			}
		}
	}
	if !valOnly { newFmt += ">>" }
	newFmt += "\n"
	return newFmt
}


// traceValRdy lists Node input values and output values or readiness.
func (n *Node) traceValRdy(valOnly bool) {

	if (!valOnly && TraceLevel<VVV || TraceLevel==Q)  {return}
	newFmt := n.traceValRdySrc(valOnly)
	newFmt += n.traceValRdyDst(valOnly)
	StdoutLog.Printf(newFmt)
}

// TraceVals lists input and output values for a Node.
func (n *Node) TraceVals() { n.traceValRdy(true) }

// IncrWorkCnt increments execution count of Node.
func (n *Node) IncrWorkCnt() {
	if (GlobalStats) {
		c := atomic.AddInt64(&globalWorkCnt, 1)
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
	n.IncrWorkCnt();
	return true
}

// Work executes Node using function pointer.
func (n *Node) Work() {
	if (n.WorkFunc!=nil) { n.WorkFunc(n) }
}


// SendAll writes all data and acks after new result is computed.
func (n *Node) SendAll() {
	for i := range n.Srcs {
		n.Srcs[i].SendAck(n)
	}
	for i := range n.Dsts {
		n.Dsts[i].SendData(n)
	}
}

// RecvOne reads one data or ack and decrements RdyCnt.
func (n *Node) RecvOne() {
	n.traceValRdy(false)
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
			newFmt := n.traceValRdySrc(true)
			n.Work()
			newFmt += n.traceValRdyDst(true)
			StdoutLog.Printf(newFmt)
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

