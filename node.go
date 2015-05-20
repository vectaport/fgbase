package flowgraph

import (
	"fmt"
	"reflect"
	"runtime"
	"strconv"
	"sync/atomic"
	"time"
)

// Node of a flowgraph.
type Node struct {
	ID int64                        // unique id
	Name string                     // for tracing
	Cnt int64                       // execution count
	Srcs []*Edge                    // upstream links
	Dsts []*Edge                    // downstream links
	RdyFunc NodeRdy                 // func to test Edge readiness
	FireFunc NodeFire               // func to fire off the Node
	RunFunc NodeRun                 // func to repeatedly run Node

	cases []reflect.SelectCase      // select cases to read from Edge's
	dataBackup []reflect.Value      // backup data channels
	caseToEdgeDir map [int] edgeDir // map from selected case to associated Edge
	flag uintptr                    // flags for package internal use
}

type edgeDir struct {
	edge *Edge
	srcFlag bool
}

const (
	flagPool = uintptr(1<<iota)
	flagRecursed
)

var startTime time.Time

// NodeRdy is the function signature for evaluating readiness of a Node to execute.
type NodeRdy func(*Node) bool

// NodeFire is the function signature for executing a Node.
// Any error message should be written using Node.LogError and
// nil written to any output Edge.
type NodeFire func(*Node)

// NodeRun is the function signature for an alternate Node event loop.
type NodeRun func(*Node)

func makeNode(name string, srcs, dsts []*Edge, ready NodeRdy, fire NodeFire, pool bool) Node {
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
	if pool { n.flag = n.flag | flagPool }
	var cnt = 0
	for i := range n.Srcs {
		srci := n.Srcs[i]
		srci.RdyCnt = func () int {
			if srci.Val!=nil { return 0 }; return 1}()
		if srci.Data != nil {
			j := len(*srci.Data)
			if j==0 || !pool {
				var df = func() int {if pool {return 0} else {return 1}}
				*srci.Data = append(*srci.Data, make(chan Datum, df()))
			} else {
				j = 0
			}
			n.cases = append(n.cases, reflect.SelectCase{Dir:reflect.SelectRecv, Chan:reflect.ValueOf((*srci.Data)[j])})
			n.dataBackup = append(n.dataBackup, n.cases[cnt].Chan)  // backup copy
			n.caseToEdgeDir[cnt] = edgeDir{srci, true}
			cnt = cnt+1
		}
	}
	for i := range n.Dsts {
		dsti := n.Dsts[i]
		dsti.RdyCnt = func (b bool) int {if b { return 0 }; return len(*dsti.Data) } (dsti.Val==nil)
		if dsti.Ack!=nil {
			if pool {
				dsti.Ack = make(chan Nada, 1)
			}
			n.cases = append(n.cases, reflect.SelectCase{Dir:reflect.SelectRecv, Chan:reflect.ValueOf(dsti.Ack)})
			n.caseToEdgeDir[cnt] = edgeDir{dsti, false}
			cnt = cnt+1
		}
	}

	return n
}

// MakeNodePool returns a new Node with slices of input and output Edge's and functions for testing readiness then firing.
// Both source channels and the destination data channel get shared.  The destination ack channel is unique.
func MakeNodePool(
	name string, 
	srcs, dsts []Edge, 
	ready NodeRdy, 
	fire NodeFire) Node {
	var srcsp,dstsp []*Edge
	for i:=0; i<len(srcs); i++ {
		srcsp = append(srcsp, &srcs[i])
	}
	for i:=0; i<len(dsts); i++ {
		dstsp = append(dstsp, &dsts[i])
	}
	return makeNode(name, srcsp, dstsp, ready, fire, true)
}

// MakeNode returns a new Node with slices of input and output Edge's and functions for testing readiness then firing.
func MakeNode(
	name string, 
	srcs, dsts []*Edge, 
	ready NodeRdy, 
	fire NodeFire) Node {
	return makeNode(name, srcs, dsts, ready, fire, false)
}

func prefixTracef(n *Node) (format string) {
	var newFmt string
	if TraceIndent {
		for i := int64(0);i<n.ID;i++ {
			newFmt += "\t"
		}
	}
	newFmt += n.Name
	newFmt += fmt.Sprintf("(%d", n.ID)

	if TraceFireCnt {
		if n.Cnt>=0 {
			newFmt += fmt.Sprintf(":%d", n.Cnt)
		} else {
			newFmt += ":*"
		}
	}

	if TraceSeconds {
		newFmt += fmt.Sprintf(":%.4f", time.Since(startTime).Seconds())
	}

	if TracePointer || TraceLevel >= VVVV {
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

// LogError for logging of error messages.  Uses atomic log mechanism.
func (n *Node) LogError(format string, v ...interface{}) {
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
		if _,ok := dstiv.(nodeWrap); ok {
			dstiv = dstiv.(nodeWrap).datum // remove wrapper for tracing
		}
		if (i!=0) { newFmt += "," }
		if (valOnly) {
			newFmt += fmt.Sprintf("%s=", dsti.Name)
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
		} else {
			if true {
				newFmt += fmt.Sprintf("%s={%v}", dsti.Name, dsti.RdyCnt)
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

// traceValRdyErr lists Node input values and output readiness to stderr.
func (n *Node) traceValRdyErr() {

	newFmt := n.traceValRdySrc(false)
	newFmt += n.traceValRdyDst(false)
	StderrLog.Printf(newFmt)
}

// TraceVals lists input and output values for a Node.
func (n *Node) TraceVals() { if TraceLevel!=Q { n.traceValRdy(true) } }

// incrFireCnt increments execution count of Node.
func (n *Node) incrFireCnt() {
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

	n.incrFireCnt();

	// restore data channels for next use
	for i := range n.dataBackup {
		n.cases[i].Chan = n.dataBackup[i]
	}

	return true
}

// Fire executes Node using function pointer.
func (n *Node) Fire() {
	var newFmt string
	if TraceLevel>Q { newFmt = n.traceValRdySrc(true) }
	if (n.FireFunc!=nil) { n.FireFunc(n) }
	if TraceLevel>Q { 
		newFmt += n.traceValRdyDst(true)
		StdoutLog.Printf(newFmt)
	}
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
		n.LogError("receive from select not ok for i=%d case", i);
		return false
	}
	if n.caseToEdgeDir[i].srcFlag {
		n.cases[i].Chan = reflect.ValueOf(nil) // don't read this again until after RdyAll
		srci := n.caseToEdgeDir[i].edge
		srci.Val = recv.Interface()
		var asterisk string
		if _,ok := srci.Val.(nodeWrap); ok {
			n2 := srci.Val.(nodeWrap).node
			srci.Ack2 = n2.Dsts[0].Ack
			srci.Val = srci.Val.(nodeWrap).datum
			if TraceLevel>=VV { asterisk = fmt.Sprintf(" *(Ack2=%p)", srci.Ack2) }
			if &n2.FireFunc == &n.FireFunc { 
				n.flag |=flagRecursed 
			} else {
				bitr := ^flagRecursed
				n.flag =(n.flag & ^bitr)
			}
		}
		srci.RdyCnt--
		if (TraceLevel>=VV) {
			if (srci.Val==nil) {
				n.Tracef("<nil> <- %s.Data%s\n", srci.Name, asterisk)
			} else {
				n.Tracef("%s <- %s.Data%s\n", String(srci.Val), srci.Name, asterisk)
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
			if TraceLevel >= VVV {n.traceValRdy(false)}
			n.Fire()
			n.SendAll()
		}
		if !n.RecvOne() { // bad receiving shuts down go-routine
			break
		}
	}
}

// MakeNodes returns a slice of Node.
func MakeNodes(sz int) []Node {
	n := make([]Node, sz)
	return n
}

// RunAll calls Run for each Node.
func RunAll(n []Node, timeout time.Duration) {
		
	startTime = time.Now()
	for i:=0; i<len(n); i++ {
		var node *Node = &n[i]
		if TraceLevel>=VVVV {
			node.Tracef("\n")
		}
		go node.Run()
	}

	if timeout>0 { 
		time.Sleep(timeout) 
		defer StdoutLog.Printf("\n")
	}

	if PostDump {
		StderrLog.Printf(">>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>\n")
		for i:=0; i<len(n); i++ {
			n[i].traceValRdyErr()
		}
		StderrLog.Printf("<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<\n")
	}
}

// NodeWrap bundles a Node pointer and a Datum to pass information about an
// upstream node downstream.  Used for acking back in a Pool.
func (n *Node) NodeWrap(d Datum) Datum {
	return nodeWrap{n, d}
}

// Recursed returns true if a Node from the same Pool is upstream of this Node.
func (n *Node) Recursed() bool { return n.flag&flagRecursed==flagRecursed }

// IsPool returns true if Node is part of a Pool.
func (n *Node) IsPool() bool { return n.flag&flagPool==flagPool }
