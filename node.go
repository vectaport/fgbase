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
	dataBackup []reflect.Value      // backup data channels
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

func makeNode(name string, srcs, dsts []*Edge, ready NodeRdy, work NodeWork, reuseChan bool) Node {
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
			if j==0 || !reuseChan {
				*n.Srcs[i].Data = append(*n.Srcs[i].Data, make(chan Datum, 0))
			} else {
				j = 0
			}
			n.cases = append(n.cases, reflect.SelectCase{Dir:reflect.SelectRecv, Chan:reflect.ValueOf((*n.Srcs[i].Data)[j])})
			n.dataBackup = append(n.dataBackup, n.cases[cnt].Chan)  // backup copy
			n.caseToEdgeDir[cnt] = edgeDir{n.Srcs[i], true}
			cnt = cnt+1
		}
	}
	for i := range n.Dsts {
		n.Dsts[i].RdyCnt = func (b bool) int {if b { return 0 }; return len(*n.Dsts[i].Data) } (n.Dsts[i].Val==nil)
		if n.Dsts[i].Ack!=nil {
			if reuseChan {
				n.Dsts[i].Ack = make(chan bool, 0)
			}
			n.cases = append(n.cases, reflect.SelectCase{Dir:reflect.SelectRecv, Chan:reflect.ValueOf(n.Dsts[i].Ack)})
			n.caseToEdgeDir[cnt] = edgeDir{n.Dsts[i], false}
			cnt = cnt+1
		}
	}

	return n
}

// MakeNode2 returns a new Node with slices of input and output Edge's and functions for testing readiness then firing.
// The Edge data channels get reused.
func MakeNode2(
	name string, 
	srcs, dsts []*Edge, 
	ready NodeRdy, 
	work NodeWork) Node {
	return makeNode(name, srcs, dsts, ready, work, true)
}

// MakeNode returns a new Node with slices of input and output Edge's and functions for testing readiness then firing.
func MakeNode(
	name string, 
	srcs, dsts []*Edge, 
	ready NodeRdy, 
	work NodeWork) Node {
	return makeNode(name, srcs, dsts, ready, work, false)
}

func prefixTracef(n *Node) (format string) {
	var addNodeAddr = TraceLevel>=VVVV
	var newFmt string
	if (TraceIndent) {
		for i := int64(0);i<n.ID;i++ {
			newFmt += "\t"
		}
	}
	newFmt += n.Name
	newFmt += fmt.Sprintf("(%d", n.ID)
	if (n.Cnt>=0) {
		newFmt += fmt.Sprintf(":%d", n.Cnt)
	} else {
		newFmt += ":*"
	}
	if (addNodeAddr) { 
		newFmt += fmt.Sprintf(":%p", n)
	}
	newFmt += ") "
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
				newFmt +=StringSlice(srci.Val)
			} else {
				if srci.Val==nil  {
					newFmt += "<nil>"
				} else {
					newFmt += String(srci.Val)
				}
			}
		} else {
			newFmt += "{}"
		}
	}
	newFmt += ";"
	return newFmt
}

// traceValRdyDst lists Node output values or readiness.
func (n *Node) traceValRdyDst(valOnly bool) string {
	var newFmt string
	for i := range n.Dsts {
		dsti := n.Dsts[i]
		dstiv := dsti.Val
		if _,ok := dstiv.(ackWrap); ok {
			dstiv = dstiv.(ackWrap).d // remove wrapper for tracing
		}
		if (i!=0) { newFmt += "," }
		if (valOnly) {
			newFmt += fmt.Sprintf("%s=", dsti.Name)
			if IsSlice(dstiv) {
				newFmt += StringSlice(dstiv)
			} else {
				if (dstiv != nil) {
					newFmt += String(dstiv)
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

	newFmt := n.traceValRdySrc(valOnly)
	newFmt += n.traceValRdyDst(valOnly)
	StdoutLog.Printf(newFmt)
}

// TraceVals lists input and output values for a Node.
func (n *Node) TraceVals() { if TraceLevel!=Q { n.traceValRdy(true) } }

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

	// restore data channels for next use
	for i := range n.dataBackup {
		n.cases[i].Chan = n.dataBackup[i]
	}

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
func (n *Node) RecvOne() (recvOK bool) {
	if TraceLevel >= VVV {n.traceValRdy(false)}
	i,recv,recvOK := reflect.Select(n.cases)
	if !recvOK {
		n.Errorf("receive not ok for i=%d case\n", i);
		return false
	}
	if n.caseToEdgeDir[i].srcFlag {
		n.cases[i].Chan = reflect.ValueOf(nil) // don't read this again until after RdyAll
		srci := n.caseToEdgeDir[i].edge
		srci.Val = recv.Interface()
		var asterisk string
		if _,ok := srci.Val.(ackWrap); ok {
			srci.Ack2 = srci.Val.(ackWrap).ack
			srci.Val = srci.Val.(ackWrap).d
			asterisk = fmt.Sprintf(" *(Ack2=%p)", srci.Ack2)
		}
		srci.RdyCnt--
		if (TraceLevel>=VV) {
			if (srci.Val==nil) {
				n.Tracef("<nil> <- %s.Data%s\n", srci.Name, asterisk)
			} else {
				n.Tracef("%s <- %s.Data%s\n", 
					func() string {
						if IsSlice(srci.Val) { return StringSlice(srci.Val) }
						return fmt.Sprintf("%T(%v)", srci.Val, srci.Val)}(), 
					srci.Name,
					asterisk)
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
			n.Tracef("<- %s(%p)\n", nm, dsti.Ack)
		}
	}
	return recvOK
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
		if !n.RecvOne() { // bad receiving shutsdown go-routine
			break
		}
	}
}

// MakeNodes returns a slice of Node.
func MakeNodes(sz int) []Node {
	n := make([]Node, sz)
	return n
}
