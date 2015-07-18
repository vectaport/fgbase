package flowgraph

import (
	"fmt"
	"reflect"
	"runtime"
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
	caseToEdgeDir map [int] edgeDir // map from index of selected case to associated Edge
	edgeToCase map [*Edge] int      // map from *Edge to index of associated select case
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

// NodeRdy is the function signature for evaluating readiness of a Node to fire.
type NodeRdy func(*Node) bool

// NodeFire is the function signature for executing a Node.
// Any error message should be written using Node.LogError and
// nil written to any output Edge.
type NodeFire func(*Node)

// NodeRun is the function signature for an alternate Node event loop.
type NodeRun func(*Node)

func makeNode(name string, srcs, dsts []*Edge, ready NodeRdy, fire NodeFire, pool, recurse bool) Node {
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
	n.edgeToCase = make(map[*Edge]int)
	if pool { n.flag = n.flag | flagPool }
	var cnt = 0
	for i := range n.Srcs {
		srci := n.Srcs[i]
		srci.RdyCnt = func () int {
			if srci.Val!=nil { return 0 }; return 1}()
		if srci.Data != nil {
			j := len(*srci.Data)
			if j==0 || !pool {
				var df = func() int {if pool&&recurse {return 0}; return ChannelSize}
				*srci.Data = append(*srci.Data, make(chan Datum, df()))
			} else {
				j = 0
			}
			n.cases = append(n.cases, reflect.SelectCase{Dir:reflect.SelectRecv, Chan:reflect.ValueOf((*srci.Data)[j])})
			n.dataBackup = append(n.dataBackup, n.cases[cnt].Chan)  // backup copy
			n.caseToEdgeDir[cnt] = edgeDir{srci, true}
			n.edgeToCase[srci] = cnt
			cnt = cnt+1
		}
	}
	for i := range n.Dsts {
		dsti := n.Dsts[i]
		dsti.RdyCnt = func (b bool) int {if b { return 0 }; return len(*dsti.Data) } (dsti.Val==nil)
		if dsti.Ack!=nil {
			if pool {
				dsti.Ack = make(chan Nada, ChannelSize)
			}
			n.cases = append(n.cases, reflect.SelectCase{Dir:reflect.SelectRecv, Chan:reflect.ValueOf(dsti.Ack)})
			n.caseToEdgeDir[cnt] = edgeDir{dsti, false}
			n.edgeToCase[dsti] = cnt
			cnt = cnt+1
		}
	}

	return n
}

// makeNodeForPool returns a new Node with copies of source and destination Edge's.
// Both source channels and the destination data channel get shared.  
// The destination ack channel is unique.
func makeNodeForPool(
	name string, 
	srcs, dsts []Edge, 
	ready NodeRdy, 
	fire NodeFire,
        recurse bool) Node {
	var srcsp,dstsp []*Edge
	for i:=0; i<len(srcs); i++ {
		srcsp = append(srcsp, &srcs[i])
	}
	for i:=0; i<len(dsts); i++ {
		dstsp = append(dstsp, &dsts[i])
	}
	return makeNode(name, srcsp, dstsp, ready, fire, true, recurse)
}

// MakeNode returns a new Node with slices of input and output Edge's and functions for testing readiness then firing.
func MakeNode(
	name string, 
	srcs, dsts []*Edge, 
	ready NodeRdy, 
	fire NodeFire) Node {
	return makeNode(name, srcs, dsts, ready, fire, false, false)
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

	if TraceSeconds  || TraceLevel >= VVVV {
		newFmt += fmt.Sprintf(":%.4f", TimeSinceStart())
	}

	if TracePointer{
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
					newFmt += fmt.Sprintf("%s", String(srci.Val))
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
				s := String(dstiv)
				if !IsSlice(dstiv) {
					newFmt += fmt.Sprintf("%s", s)
				} else {
					newFmt += s
				}

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

// TraceValRdy lists Node input values and output readiness when TraceLevel >= VVV.
func (n *Node) TraceValRdy() {
	if TraceLevel >= VVV {
		n.traceValRdy(false)
	}
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
			if !n.Srcs[i].SrcRdy(n) {
				return false
			}
		}
		for i := range n.Dsts {
			if !n.Dsts[i].DstRdy(n) {
				return false
			}
		}
	} else {
		if !n.RdyFunc(n) { 
			return false 
		}
	}
	
	// restore data channels for next use
	for i := range n.dataBackup {
		n.cases[i].Chan = n.dataBackup[i]
	}
	
	return true
}

// Fire executes Node using function pointer.
func (n *Node) Fire() {
	n.incrFireCnt();
	var newFmt string
	if TraceLevel>Q { newFmt = n.traceValRdySrc(true) }
	if (n.FireFunc!=nil) { 
		n.FireFunc(n) 
	} 
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
		srci := n.caseToEdgeDir[i].edge
		srci.Val = recv.Interface()
		n.cases[i].Chan = reflect.ValueOf(nil) // don't read this again until after RdyAll
		srci.srcReadHandle(n, true)
	} else {
		dsti := n.caseToEdgeDir[i].edge
		dsti.dstReadHandle(n, true)
	}
	return recvOK
}

// DefaultRunFunc is the default run func.
func (n *Node) DefaultRunFunc () {
	for {
		for n.RdyAll() {
			if TraceLevel >= VVV {n.traceValRdy(false)}
			n.Fire()
			n.SendAll()
		}
		if !n.RecvOne() { // bad receiving shuts down go-routine
			break
		}
	}
}

// Run is an event loop that runs forever for each Node.
func (n *Node) Run() {
	if n.RunFunc != nil {
		n.RunFunc(n)
		return
	}

	n.DefaultRunFunc()
}

// FireThenWait fires off a ready Node then waits until it is ready again.
func (n *Node) FireThenWait() {

	if TraceLevel >= VVV {n.traceValRdy(false)}
	if n.RdyAll() {
		n.Fire()
		n.SendAll()
	}

}


// MakeNodes returns a slice of Node.
func MakeNodes(sz int) []Node {
	n := make([]Node, sz)
	return n
}

// RunAll calls Run for each Node, and timesout after RunTime.
func RunAll(n []Node) {
		
	StartTime = time.Now()
	for i:=0; i<len(n); i++ {
		var node = &n[i]
		if TraceLevel>=VVVV {
			node.Tracef("\n")
		}
		go node.Run()
	}

	timeout := RunTime
	if timeout>0 { 
		time.Sleep(timeout) 
		defer StdoutLog.Printf("\n")
	}

	if TraceLevel>=VVVV {
		StdoutLog.Printf("<<<<<<<<<<<<<<<<>>>>>>>>>>>>>>>>\n")
		for i:=0; i<len(n); i++ {
			n[i].traceValRdy(false)
		}
	}
}

// NodeWrap bundles a Node pointer, and an ack channel with a Datum, in order to 
// pass information about an upstream node downstream.  Used for acking back in a Pool.
func (n *Node) NodeWrap(d Datum, ack chan Nada) Datum {
	return nodeWrap{n, d, ack}
}

// Recursed returns true if a Node from the same Pool is upstream of this Node.
func (n *Node) Recursed() bool { return n.flag&flagRecursed==flagRecursed }

// IsPool returns true if Node is part of a Pool.
func (n *Node) IsPool() bool { return n.flag&flagPool==flagPool }
