package fgbase

import (
	"fmt"
	"reflect"
	"sync"
	"sync/atomic"
	"time"
)

// Node of a flowgraph.
type Node struct {
	ID       int64       // unique id
	Name     string      // for tracing
	Cnt      int64       // execution count
	Srcs     []*Edge     // upstream Edge's
	Dsts     []*Edge     // downstream Edge's
	RdyFunc  NodeRdy     // func to test Edge readiness
	FireFunc NodeFire    // func to fire off the Node
	RunFunc  NodeRun     // func to repeatedly run Node
	Aux      interface{} // auxiliary empty interface to hold state
	RdyState int         // state of latest readiness

	cases         []reflect.SelectCase // select cases to read from Edge's
	caseToEdgeDir map[int]edgeDir      // map from index of selected case to associated Edge
	edgeToCase    map[*Edge]int        // map from *Edge to index of associated select case
	dataBackup    []reflect.Value      // backup data channels for inputs
	flag          uintptr              // flags for package internal use
	srcByName     map[string]*Edge     // map of upstream Edge's by name
	dstByName     map[string]*Edge     // map of downstream Edge's by name
	srcNames      []string             // source names
	dstNames      []string             // destination names
}

type edgeDir struct {
	edge    *Edge
	srcFlag bool
}

const (
	flagPool = uintptr(1 << iota)
	flagRecurse
	flagRecursed
)

// NodeRdy is the function signature for evaluating readiness of a Node to fire.
type NodeRdy func(*Node) bool

// NodeFire is the function signature for executing a Node.
// Any error message should be written using Node.LogError and
// nil written to any output Edge.
type NodeFire func(*Node) error

// NodeRun is the function signature for an alternate Node event loop.
type NodeRun func(*Node) error

func newNode(name string, ready NodeRdy, fire NodeFire) Node {
	var n Node
	i := atomic.AddInt64(&NodeID, 1)
	n.ID = i - 1
	n.Name = name
	n.Cnt = -1
	n.RdyFunc = ready
	n.FireFunc = fire
	n.caseToEdgeDir = make(map[int]edgeDir)
	n.edgeToCase = make(map[*Edge]int)
	return n
}

func makeNode(name string, srcs, dsts []*Edge, ready NodeRdy, fire NodeFire, pool, recurse bool) Node {
	n := newNode(name, ready, fire)

	if pool {
		n.flag = n.flag | flagPool
	}
	if recurse {
		n.flag = n.flag | flagRecurse
	}

	n.Srcs = srcs
	n.Dsts = dsts

	// n.Init()

	return n

}

// Init initializes node internals after edges have been added
func (n *Node) Init() {
	var cnt = 0
	pool := (n.flag & flagPool) == flagPool
	recurse := (n.flag & flagRecurse) == flagRecurse
	for i := range n.Srcs {
		srci := n.Srcs[i]
		srci.RdyCnt = func() int {
			if srci.Val != nil {
				return 0
			}
			return 1
		}()
		if srci.Data != nil {
			j := len(*srci.Data)
			if j == 0 || !pool {
				df := func() int {
					if pool && recurse {
						return 0
					}
					return ChannelSize
				}()
				*srci.Data = append(*srci.Data, make(chan interface{}, df))
			} else {
				j = 0
			}
			n.cases = append(n.cases, reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf((*srci.Data)[j])})
			n.dataBackup = append(n.dataBackup, n.cases[cnt].Chan) // backup copy
			n.caseToEdgeDir[cnt] = edgeDir{srci, true}
			n.edgeToCase[srci] = cnt
			cnt = cnt + 1
		}
	}
	for i := range n.Dsts {
		dsti := n.Dsts[i]
		dsti.RdyCnt = 0

		if dsti.Ack != nil {
			if pool {
				dsti.Ack = make(chan struct{}, ChannelSize)
			}
			n.cases = append(n.cases, reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(dsti.Ack)})
			n.caseToEdgeDir[cnt] = edgeDir{dsti, false}
			n.edgeToCase[dsti] = cnt
			cnt = cnt + 1
		}
	}
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
	var srcsp, dstsp []*Edge
	for i := 0; i < len(srcs); i++ {
		srcsp = append(srcsp, &srcs[i])
	}
	for i := 0; i < len(dsts); i++ {
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
		for i := int64(0); i < n.ID; i++ {
			newFmt += "\t"
		}
	}
	newFmt += n.Name
	newFmt += fmt.Sprintf("(%d", n.ID)

	if TraceFireCnt {
		if n.Cnt >= 0 {
			newFmt += fmt.Sprintf(":%d", n.Cnt)
		} else {
			newFmt += ":*"
		}
	}

	if TraceSeconds || TraceLevel >= VVVV {
		t := TimeSinceStart()
		if t >= 0.0 {
			newFmt += fmt.Sprintf(":%.4f", TimeSinceStart())
		} else {
			newFmt += ":*"
		}
	}

	if TracePointer {
		newFmt += fmt.Sprintf(":%p", n)
	}

	newFmt += ") "

	return newFmt
}

// Tracef for debug trace printing.  Uses atomic log mechanism.
func (n *Node) Tracef(format string, v ...interface{}) {
	if TraceLevel < V {
		return
	}
	newFmt := prefixTracef(n)
	newFmt += format
	StdoutLog.Printf(newFmt, v...)
}

// LogError for logging of error messages.  Uses atomic log mechanism.
func (n *Node) LogError(format string, v ...interface{}) {
	// _,nm,ln,_ := runtime.Caller(1)
	newFmt := prefixTracef(n)
	newFmt += "ERROR:  "
	newFmt += format
	// newFmt += fmt.Sprintf(" -- %s:%d ", nm, ln)
	StderrLog.Printf(newFmt, v...)
}

// traceValRdySrc lists Node input values
func (n *Node) traceValRdySrc(valOnly bool) string {
	newFmt := prefixTracef(n)
	if !valOnly {
		newFmt += "<<"
	}
	for i := range n.Srcs {
		srci := n.Srcs[i]
		if i != 0 {
			newFmt += ","
		}
		newFmt += fmt.Sprintf("%s=", srci.Name)
		if srci.RdyZero() {
			if IsSlice(srci.Val) {
				newFmt += StringSlice(srci.Val)
			} else {
				if srci.Val == nil {
					newFmt += "<nil>"
				} else if v, ok := srci.Val.(error); ok && v.Error() == "EOF" {
					newFmt += "EOF"
				} else {
					newFmt += fmt.Sprintf("%s", String(srci.Val))
				}
			}
		} else {
			newFmt += "_"
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
		if _, ok := dstiv.(nodeWrap); ok {
			dstiv = dstiv.(nodeWrap).datum // remove wrapper for tracing
		}
		if i != 0 {
			newFmt += ","
		}
		if valOnly {
			newFmt += fmt.Sprintf("%s=", dsti.Name)
			if !dsti.Flow {
				newFmt += "_"
			} else {
				if dstiv != nil {
					s := String(dstiv)
					if v, ok := dstiv.(error); ok && v.Error() == "EOF" {
						newFmt += "EOF"
					} else if !IsSlice(dstiv) {
						newFmt += fmt.Sprintf("%s", s)
					} else {
						newFmt += s
					}

				} else {
					newFmt += "<nil>"
				}
			}

		} else {
			newFmt += fmt.Sprintf("%s=k%v", dsti.Name, dsti.RdyCnt)
		}
	}
	if !valOnly {
		newFmt += ">>"
	}
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
	newFmt += "\n"
	StdoutLog.Printf(newFmt)
}

// traceValRdyErr lists Node input values and output readiness to stderr.
func (n *Node) traceValRdyErr() {

	newFmt := n.traceValRdySrc(false)
	newFmt += n.traceValRdyDst(false)
	newFmt += "\n"
	StderrLog.Printf(newFmt)
}

// TraceVals lists input and output values for a Node.
func (n *Node) TraceVals() {
	if TraceLevel != Q {
		n.traceValRdy(true)
	}
}

// incrFireCnt increments execution count of Node.
func (n *Node) incrFireCnt() {
	if GlobalStats {
		c := atomic.AddInt64(&globalFireCnt, 1)
		n.Cnt = c - 1
	} else {
		n.Cnt = n.Cnt + 1
	}
}

// DefaultRdyFunc tests for everything ready.
func (n *Node) DefaultRdyFunc() bool {
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
	return true
}

// restoreDataChannels restore data channels for next use
func (n *Node) restoreDataChannels() {
	for i := range n.dataBackup {
		if n.Srcs[i].RdyCnt > 0 {
			n.cases[i].Chan = n.dataBackup[i]
		}
	}
}

// RestoreDataChannel restores the data channel for a given source edge
func (n *Node) RestoreDataChannel(e *Edge) {
	for i := range n.Srcs {
		if n.Srcs[i] == e {
			n.cases[i].Chan = n.dataBackup[i]
			break
		}
	}
}

// RdyAll tests readiness of Node to execute.
func (n *Node) RdyAll() bool {

	if n.RdyFunc == nil {
		if !n.DefaultRdyFunc() {
			return false
		}
	} else {
		if !n.RdyFunc(n) {
			return false
		}
	}

	return true
}

// Fire executes Node using function pointer.
func (n *Node) Fire() error {
	var err error = nil
	n.incrFireCnt()
	var newFmt string
	if TraceLevel > Q {
		newFmt = n.traceValRdySrc(true)
	}
	if n.FireFunc != nil {
		err = n.FireFunc(n)
	} else {
		for i := range n.Srcs {
			n.Srcs[i].Flow = true
		}
		for i := range n.Dsts {
			n.Dsts[i].Flow = true
		}
	}
	if TraceLevel > Q {
		// newFmt += "\t"
		newFmt += n.traceValRdyDst(true)
		if n.Aux != nil && IsStruct(n.Aux) {
			s := String(n.Aux)
			if s != "{}" {
				// newFmt += "\t// " + s
				newFmt += " // " + s
			}
		}
		newFmt += "\n"
		StdoutLog.Printf(newFmt)
	}
	return err
}

// SendAll writes all data and acks after new result is computed.
func (n *Node) SendAll() bool {
	sent := false
	for i := range n.Srcs {
		sent = n.Srcs[i].SendAck(n) || sent
	}
	for i := range n.Dsts {
		sent = n.Dsts[i].SendData(n) || sent
	}
	return sent
}

// RecvOne reads one data or ack and decrements RdyCnt.
func (n *Node) RecvOne() (recvOK bool) {
	if TraceLevel >= VVV {
		n.traceValRdy(false)
	}
	i, recv, recvOK := reflect.Select(n.cases)
	if !recvOK {
		n.LogError("receive from select not ok for case %d", i)
		return false
	}
	if n.caseToEdgeDir[i].srcFlag {
		srci := n.caseToEdgeDir[i].edge
		srci.Val = recv.Interface()
		n.RemoveInputCase(srci)
		srci.srcReadHandle(n, true)
	} else {
		dsti := n.caseToEdgeDir[i].edge
		dsti.dstReadHandle(n, true)
	}
	return recvOK
}

// DefaultRunFunc is the default run func.
func (n *Node) DefaultRunFunc() error {
	var err error = nil
	for {
		for n.RdyAll() {
			if TraceLevel >= VVV {
				n.traceValRdy(false)
			}
			err = n.Fire()
			sent := n.SendAll()
			n.restoreDataChannels()
			if err != nil {
				return err
			}
			if !sent {
				break
			} // wait for external event

		}
		if !n.RecvOne() { // bad receiving shuts down go-routine
			break
		}
	}
	return nil
}

// Run is an event loop that runs forever for each Node.
func (n *Node) Run() error {
	if n.RunFunc != nil {
		return n.RunFunc(n)
	}

	err := n.DefaultRunFunc()
	return err
}

// MakeNodes returns a slice of Node.
func MakeNodes(sz int) []Node {
	n := make([]Node, sz)
	return n
}

// buildEdgeNodes builds the slice of edgeNode for each Edge to allow reflection.
func buildEdgeNodes(nodes []Node) {
	for i, n := range nodes {
		for j := range n.Srcs {
			srcj := n.Srcs[j]
			        if srcj.edgeNodes==nil {
					   panic("Using an Edge that has an uninitialized edgeNodes\n")
					   	   }
			*srcj.edgeNodes = append(*srcj.edgeNodes, edgeNode{node: &nodes[i], srcFlag: false})
		}
		for j := range n.Dsts {
			dstj := n.Dsts[j]
			k := 0
			for ; k < len(*dstj.edgeNodes) && (*dstj.edgeNodes)[k].srcFlag; k++ {
			}
			*dstj.edgeNodes = append(*dstj.edgeNodes, edgeNode{})
			copy((*dstj.edgeNodes)[k+1:], (*dstj.edgeNodes)[k:])
			(*dstj.edgeNodes)[k] = edgeNode{node: &nodes[i], srcFlag: true}
		}
	}
}

// extendChannelCaps extends the channel capacity to support arbitrated fan-in.
func extendChannelCaps(nodes []Node) {
	for _, n := range nodes {
		for j := range n.Dsts {
			dstj := n.Dsts[j]
			h := dstj.SrcCnt()
			if h > 1 {
				l := len(*dstj.Data)
				for k := 0; k < l; k++ {
					if cap((*dstj.Data)[k]) < h {
						// StdoutLog.Printf("Multiple upstream nodes on %s (len(*dstj.Data)=%d vs dstj.SrcCnt()=%d)\n", dstj.Name, len(*dstj.Data), dstj.DstCnt())
						c := make(chan interface{}, h)
						(*dstj.Data)[k] = c

						// update relevant select case and data channel upstream
						nn := dstj.DstNode(0)
						x := nn.edgeToCase[nn.Srcs[0]]
						nn.cases[x] = reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(c)}
						nn.dataBackup[x] = nn.cases[x].Chan

					}

				}
			}
			if false {
				if len(*dstj.Data) != dstj.DstCnt() {
					StdoutLog.Printf("Multiple downstream nodes on %s (len(*dstj.Data)=%d vs dstj.DstCnt()=%d) -- %v\n", dstj.Name, len(*dstj.Data), dstj.SrcCnt(), dstj.edgeNodes)
				}
			}

		}
	}
}

// clearUpstreamAcks increments RdyCnt upstream from every initialized downstream Edge
// (Node input edge) to reflect the fact that flow is initialized here.
func clearUpstreamAcks(nodes []Node) {
	for _, n := range nodes {
		for j := range n.Srcs {
			if n.Srcs[j].Val != nil {
				n.RemoveInputCase(n.Srcs[j])
			}
		}
		for j := range n.Dsts {
			if n.Dsts[j].Val != nil {
				n.Dsts[j].RdyCnt += len(*n.Dsts[j].Data)
			}
		}
	}
}

// RunAll calls Run for each Node, and times out after RunTime.
func RunAll(nodes []Node) {

	for i, _ := range nodes {
		nodes[i].Init()
	}

	buildEdgeNodes(nodes) // build reflection capability

	if GmlOutput || DotOutput {
		if GmlOutput {
			OutputGml(nodes)
		} else {
			OutputDot(nodes)
		}
		TraceLevel = QQ
		return
	}

	extendChannelCaps(nodes)
	clearUpstreamAcks(nodes)

	if TraceLevel >= VVVV {
		for i := range nodes {
			nodes[i].TraceValRdy()
		}
		StdoutLog.Printf("<<<<<<<<<<<<<<<<<<<<>>>>>>>>>>>>>>>>>>>>\n")
	}
	StartTime = time.Now()
	var wg sync.WaitGroup
	wg.Add(len(nodes))
	for i := 0; i < len(nodes); i++ {
		node := &nodes[i]
		go func() {
			defer wg.Done()
			node.Run()
		}()
	}

	timeout := RunTime
	if timeout > 0 {
		time.Sleep(timeout)
		if TraceLevel > QQ {
			defer StdoutLog.Printf("\n")
		}
	} else {
		wg.Wait()
	}

	if TraceLevel >= VVVV {
		StdoutLog.Printf("<<<<<<<<<<<<<<<<<<<<>>>>>>>>>>>>>>>>>>>>\n")
		for i := 0; i < len(nodes); i++ {
			nodes[i].traceValRdy(false)
		}
	}

}

// NodeWrap bundles a Node pointer, and an ack channel with an empty interface, in order to
// pass information about an upstream node downstream.  Used for acking back in a Pool.
func (n *Node) NodeWrap(d interface{}, ack chan struct{}) interface{} {
	return nodeWrap{n, d, ack}
}

// Recursed returns true if a Node from the same Pool is upstream of this Node.
func (n *Node) Recursed() bool { return n.flag&flagRecursed == flagRecursed }

// IsPool returns true if Node is part of a Pool.
func (n *Node) IsPool() bool { return n.flag&flagPool == flagPool }

// RemoveInputCase removes the nth input of a Node from the select switch.
// It is restored after RdyAll.
func (n *Node) RemoveInputCase(e *Edge) {
	if !e.IsConst() {
		n.cases[n.edgeToCase[e]].Chan = reflect.ValueOf(nil) // don't read this again until after RdyAll
	}
}

// Output .dot graphviz format
func OutputDot(nodes []Node) {

	fmt.Printf("digraph G {\n")

	for _, iv := range nodes {
		for _, jv := range iv.Dsts {
			for _, kv := range *jv.edgeNodes {
				if !kv.srcFlag {
					fmt.Printf("%s_%d", iv.Name, iv.ID)
					fmt.Printf(" -> %s_%d\n", kv.node.Name, kv.node.ID)
				}
			}
		}
	}

	fmt.Printf("}\n")

}

// Output .gml graph modeling language format
func OutputGml(nodes []Node) {

	fmt.Printf("graph\n[\n")

	for _, iv := range nodes {
		fmt.Printf("  node\n  [\n   id %s_%d\n  ]\n", iv.Name, iv.ID)
	}

	for _, iv := range nodes {
		for _, jv := range iv.Dsts {
			for _, kv := range *jv.edgeNodes {
				if !kv.srcFlag {
					fmt.Printf("  edge\n  [\n   source %s_%d\n", iv.Name, iv.ID)
					fmt.Printf("   target %s_%d\n", kv.node.Name, kv.node.ID)
					fmt.Printf("   label \"%s\"\n  ]\n", jv.Name)
				}
			}
		}
	}

	fmt.Printf("]\n")

}

// SrcCnt returns the number of source edges.
func (n *Node) SrcCnt() int {
	return len(n.Srcs)
}

// DstCnt returns the number of destination edges.
func (n *Node) DstCnt() int {
	return len(n.Dsts)
}

// FindSrc returns incoming edge by name
func (n *Node) FindSrc(name string) *Edge {
	v := n.srcByName[name]
	if v != nil {
		return v
	}
	for i := range n.srcNames {
		if name == n.srcNames[i] {
			return n.Srcs[i]
		}
	}
	return nil

}

// FindDst returns incoming edge by name
func (n *Node) FindDst(name string) *Edge {
	v := n.dstByName[name]
	if v != nil {
		return v
	}
	for i := range n.dstNames {
		if name == n.dstNames[i] {
			return n.Dsts[i]
		}
	}
	return nil

}

// SetSrcNames names the incoming edges
func (n *Node) SetSrcNames(name ...string) {
	n.srcNames = name
}

// SetDstNames names the outgoing edges
func (n *Node) SetDstNames(name ...string) {
	n.dstNames = name
}

// SrcNames returns the names of the incoming edges
func (n *Node) SrcNames() []string {
	return n.srcNames
}

// DstNames returns the names of the outgoing edges
func (n *Node) DstNames() []string {
	return n.dstNames
}
